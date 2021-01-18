package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/lnsp/cloudsort/control"
	"github.com/lnsp/cloudsort/pb"
	"github.com/lnsp/cloudsort/worker"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var rootCmd = &cobra.Command{
	Use:   "cloudsort",
	Short: "cloudsort sorts gensort-formatted files",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if rootPprof != "" {
			go func() {
				if err := http.ListenAndServe(rootPprof, nil); err != nil {
					zap.S().Errorf("Start pprof handler: %s", err)
				}
			}()
		}
		return nil
	},
}

var submitCmd = &cobra.Command{
	Use:   "run",
	Short: "Submit a new sorting task",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runAndExit(func() error {
			grpcClient, err := grpc.Dial(submitControlAddr, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer grpcClient.Close()
			cc := pb.NewControlClient(grpcClient)
			jobEvents, err := cc.SubmitJob(context.Background(), &pb.SubmitJobRequest{
				Name: submitJobName,
				Creds: &pb.S3Credentials{
					Endpoints:       submitS3Endpoints,
					Region:          submitS3Region,
					AccessKeyId:     submitS3AccessKeyID,
					SecretAccessKey: submitS3SecretAccessKey,
					BucketId:        submitS3BucketID,
					ObjectKey:       submitS3ObjectKey,
					DisableSsl:      submitS3DisableSSL,
				},
			})
			if err != nil {
				return err
			}
			start := time.Now()
			first := true
			for {
				resp, err := jobEvents.Recv()
				if err == io.EOF {
					return nil
				} else if err != nil {
					return err
				}
				// Print header on first event
				if first {
					fmt.Printf("%-10s %-10s %s\n", "TIMESTAMP", "PROGRESS", "MESSAGE")
					first = false
				}
				// Print row
				event := resp.Event
				fmt.Printf("%-10.2f %-10.2f %s\n", event.Timestamp.AsTime().Sub(start).Seconds(), event.Progress, event.Message)
			}
		})
	},
}

var (
	submitJobName           string
	submitS3Endpoints       []string
	submitS3Region          string
	submitS3AccessKeyID     string
	submitS3SecretAccessKey string
	submitS3BucketID        string
	submitS3ObjectKey       string
	submitS3DisableSSL      bool
	submitControlAddr       string
)

func runAndExit(f func() error) {
	if err := f(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var controlCmd = &cobra.Command{
	Use:   "control",
	Short: "Starts up a new master instance",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runAndExit(func() error {
			listener, err := net.Listen("tcp", controlAddr)
			if err != nil {
				return err
			}
			allocatedAddr := listener.Addr().String()
			defer listener.Close()
			ctlServer, err := control.New(allocatedAddr)
			if err != nil {
				return err
			}
			grpcServer := grpc.NewServer()
			pb.RegisterControlServer(grpcServer, ctlServer)
			if err := grpcServer.Serve(listener); err != nil {
				return err
			}
			return nil
		})
	},
}

var (
	controlAddr string
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts up a new worker instance",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runAndExit(func() error {
			listener, err := net.Listen("tcp", workerAddr)
			if err != nil {
				return err
			}
			defer listener.Close()
			allocatedAddr := listener.Addr().String()
			worker, err := worker.New(allocatedAddr, workerControlAddr, workerHost, workerTotalMemory)
			if err != nil {
				return err
			}
			return worker.Wait()
		})
	},
}

var (
	workerHost        string
	workerControlAddr string
	workerAddr        string
	workerTotalMemory int
)

var (
	rootPprof string
)

func init() {
	submitFlags := submitCmd.Flags()
	submitFlags.StringVar(&submitJobName, "name", "job", "Unique name of the job")
	submitFlags.StringSliceVar(&submitS3Endpoints, "s3-endpoint", []string{"localhost:9000"}, "S3 Endpoint")
	submitFlags.StringVar(&submitS3Region, "s3-region", "us-east-1", "S3 Region")
	submitFlags.StringVar(&submitS3AccessKeyID, "s3-access-key", "minioadmin", "S3 Access Key ID")
	submitFlags.StringVar(&submitS3SecretAccessKey, "s3-secret-key", "minioadmin", "S3 Secret Key")
	submitFlags.StringVar(&submitS3BucketID, "s3-bucket-id", "cbdp-test", "S3 Bucket ID")
	submitFlags.StringVar(&submitS3ObjectKey, "s3-object-key", "test-file", "S3 Object Key")
	submitFlags.BoolVar(&submitS3DisableSSL, "s3-disable-ssl", true, "S3 disable SSL option")
	submitFlags.StringVar(&submitControlAddr, "control", "localhost:6000", "Control server address")
	controlFlags := controlCmd.Flags()
	controlFlags.StringVar(&controlAddr, "addr", "localhost:6000", "Address to listen on")
	workerFlags := workerCmd.Flags()
	workerFlags.StringVar(&workerAddr, "addr", "localhost:0", "Address to use for gRPC")
	workerFlags.StringVar(&workerControlAddr, "control", "localhost:6000", "Control server address")
	workerFlags.StringVar(&workerHost, "host", "localhost", "Host to use for raw TCP")
	workerFlags.IntVar(&workerTotalMemory, "memory", 4_000_000_000, "max memory alloc during task execution")
	rootCmd.PersistentFlags().StringVar(&rootPprof, "pprof", "", "enable pprof on address")
	rootCmd.AddCommand(submitCmd, controlCmd, workerCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
