package main

import (
	"context"
	"errors"
	"net"
	"net/http"

	"gitea.xscloud.ru/xscloud/golib/pkg/application/logging"
	libio "gitea.xscloud.ru/xscloud/golib/pkg/common/io"
	libamqp "gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/amqp"
	"gitea.xscloud.ru/xscloud/golib/pkg/infrastructure/mysql"
	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order/api/server/orderinternal"
	appservice "order/pkg/application/service"
	infraamqp "order/pkg/infrastructure/amqp"
	"order/pkg/infrastructure/client"
	inframysql "order/pkg/infrastructure/mysql"
	"order/pkg/infrastructure/mysql/query"
	infratemporal "order/pkg/infrastructure/temporal"
	"order/pkg/infrastructure/transport"
	"order/pkg/infrastructure/transport/middlewares"
)

type serviceConfig struct {
	Service  Service  `envconfig:"service"`
	Database Database `envconfig:"database" required:"true"`
	AMQP     AMQP     `envconfig:"amqp" required:"true"`
}

func service(logger logging.Logger) *cli.Command {
	return &cli.Command{
		Name:   "service",
		Before: migrateImpl(logger),
		Action: func(c *cli.Context) error {
			cnf, err := parseEnvs[serviceConfig]()
			if err != nil {
				return err
			}

			closer := libio.NewMultiCloser()
			defer func() {
				err = errors.Join(err, closer.Close())
			}()

			databaseConnector, err := newDatabaseConnector(cnf.Database)
			if err != nil {
				return err
			}
			closer.AddCloser(databaseConnector)
			databaseConnectionPool := mysql.NewConnectionPool(databaseConnector.TransactionalClient())

			libUoW := mysql.NewUnitOfWork(databaseConnectionPool, inframysql.NewRepositoryProvider)
			uow := inframysql.NewUnitOfWork(libUoW)

			productConn, err := grpc.NewClient(
				cnf.Service.ProductServiceAddress,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			closer.AddCloser(productConn)

			productClient := client.NewProductClient(productConn)

			paymentConn, err := grpc.NewClient(
				cnf.Service.PaymentServiceAddress,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			closer.AddCloser(paymentConn)

			paymentClient := client.NewPaymentClient(paymentConn)

			notificationConn, err := grpc.NewClient(
				cnf.Service.NotificationServiceAddress,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return err
			}
			closer.AddCloser(notificationConn)

			notificationClient := client.NewNotificationClient(notificationConn)

			amqpConnection := newAMQPConnection(cnf.AMQP, logger)
			amqpProducer := amqpConnection.Producer(
				&libamqp.ExchangeConfig{
					Name:    "domain_events",
					Kind:    "topic",
					Durable: true,
				},
				nil,
				nil,
			)
			err = amqpConnection.Start()
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				return amqpConnection.Stop()
			}))

			eventPublisher := infraamqp.NewEventPublisher(amqpProducer)

			// Temporal Setup
			temporalClient, err := temporalclient.Dial(temporalclient.Options{
				HostPort: cnf.Service.TemporalAddress,
			})
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				temporalClient.Close()
				return nil
			}))

			workflowStarter := infratemporal.NewWorkflowStarter(temporalClient)

			activities := infratemporal.NewActivities(uow, productClient, paymentClient, notificationClient)

			w := worker.New(temporalClient, infratemporal.TaskQueue, worker.Options{})
			w.RegisterWorkflow(infratemporal.CreateOrderWorkflow)
			w.RegisterActivity(activities)

			err = w.Start()
			if err != nil {
				return err
			}
			closer.AddCloser(libio.CloserFunc(func() error {
				w.Stop()
				return nil
			}))

			orderInternalAPI := transport.NewOrderInternalAPI(
				query.NewOrderQueryService(databaseConnector.TransactionalClient()),
				appservice.NewOrderService(uow, productClient, eventPublisher, workflowStarter),
			)

			errGroup := errgroup.Group{}
			errGroup.Go(func() error {
				listener, err := net.Listen("tcp", cnf.Service.GRPCAddress)
				if err != nil {
					return err
				}
				grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
					middlewares.NewGRPCLoggingMiddleware(logger),
				))
				orderinternal.RegisterOrderInternalServiceServer(grpcServer, orderInternalAPI)
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, func(_ context.Context) error {
					grpcServer.GracefulStop()
					return nil
				})
				return grpcServer.Serve(listener)
			})
			errGroup.Go(func() error {
				router := mux.NewRouter()
				registerHealthcheck(router)
				// nolint:gosec
				server := http.Server{
					Addr:    cnf.Service.HTTPAddress,
					Handler: router,
				}
				graceCallback(c.Context, logger, cnf.Service.GracePeriod, server.Shutdown)
				return server.ListenAndServe()
			})

			return errGroup.Wait()
		},
	}
}
