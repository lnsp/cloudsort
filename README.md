# cloudsort

This repository contains all project material required to perform the benchmark.

## Installation

```bash
$ git clone https://github.com/lnsp/cloudsort.git
$ cd cloudsort
$ make
```

## Usage

> You need a valid S3 configuration to be able to run a sort job. If you want to test it locally, please use a local `minio` instance.

### Start a control server

```bash
$ ./cloudsort control
```

### Start one or more worker instances

```bash
# Make sure that you choose a different address for each worker
$ ./cloudsort worker
```

### Submit a sort job
```
$ ./cloudsort run --control localhost:6000 --s3-endpoint localhost:9000 --s3-bucket-id cloudsort --s3-object-key gensort-10G
```

## Architecture

The project is composed of one control node and N worker nodes. The control node takes care of
- receiving new job from a client
- decomposing a job into tasks and handing them out to workers
- tracking job and task state

A typical run consists of
1. starting a control node and worker nodes
2. submitting a new job to the control node
3. the control node looks at the file size and assigns a task composed of
    - key range which gets shuffled to the worker
    - byte range which the worker should download, sort and shuffle to its peers
4. the worker performs a regular heartbeat and pulls new tasks
5. after receiving a new task, the worker downloads its chunk in multiple segment with maximum size defined by the `--mem` flag (DOWNLOAD state)
6. once a segment is finished, the worker starts to sort it immediately
7. after all segments are downloaded, they are merged back together and an index file is generated (SORT state)
8. the worker contacts all its peer to allocate a new TCP port for receiving shuffled chunks (SHUFFLE state)
9. the worker starts reading at N different positions in the merged chunk, sending each peer the respective data until a reader encounters a key which is outside of the target peer's range
10. the data received from peers is immediately merged to disk (MERGE state, however its actually concurrent with SHUFFLE mostly)
11. and then uploaded to S3 (UPLOAD state)
12. after all report finished task back to control node (DONE state)

## Project layout

| File                     | Description                                                                                       |
| ------------------------ | ------------------------------------------------------------------------------------------------- |
| `pkg/worker/worker.go`   | Worker code, responsible for pulling and running new tasks                                        |
| `pkg/worker/task.go`     | Task code, executes the different stages like downloading, merging, shuffling and uploading       |
| `pkg/worker/sort.go`     | External sort code, most of it is unused due to experimentation with different sorting approaches |
| `pkg/control/control.go` | Control code, hands out tasks to workers and broadcasts events to the job client                  |
| `pb/pb.go`               | gRPC API used for communication between client, control and worker except for shuffling           |

## Optimizations and benchmarks

- I experimented with different sorting approaches and algorithms (first sort.Slice, then TimSort now a parallel merge sort) as well as ways to compare the 10 byte keys (bytes.Compare, using Cgo and finally converting a key into one uint64 and uint16 and comparing those)
- I initially used gRPC for everything (including chunk shuffling), now its done using bare TCP connections reducing the amount of memory allocations by only using a set of two buffers (one active, one backup which are swapped when a chunk is sent/received)

## Performance

### Sorting 1TB with 10 workers and 1 control (8 vCPUs, 32GB RAM, 240GB SSD)

```
TIMESTAMP  PROGRESS   MESSAGE
0.00       0.00       Job scheduled
0.02       0.00       Target file has 1.0 TB of data
0.30       0.00       state changed 10.0.0.15:6000=DOWNLOAD
1.26       0.00       state changed 10.0.0.16:6000=DOWNLOAD
3.03       0.00       state changed 10.0.0.12:6000=DOWNLOAD
3.90       0.00       state changed 10.0.0.13:6000=DOWNLOAD
4.79       0.00       state changed 10.0.0.17:6000=DOWNLOAD
5.69       0.00       state changed 10.0.0.18:6000=DOWNLOAD
6.68       0.00       state changed 10.0.0.19:6000=DOWNLOAD
7.23       0.00       state changed 10.0.0.11:6000=DOWNLOAD
7.57       0.00       state changed 10.0.0.8:6000=DOWNLOAD
9.40       0.00       state changed 10.0.0.14:6000=DOWNLOAD
961.02     0.00       state changed 10.0.0.14:6000=SORT
988.41     0.00       state changed 10.0.0.13:6000=SORT
1000.99    0.00       state changed 10.0.0.15:6000=SORT
1132.09    0.00       state changed 10.0.0.17:6000=SORT
1283.86    0.00       state changed 10.0.0.14:6000=SHUFFLE
1302.89    0.00       state changed 10.0.0.16:6000=SORT
1308.28    0.00       state changed 10.0.0.13:6000=SHUFFLE
1320.12    0.00       state changed 10.0.0.15:6000=SHUFFLE
1339.71    0.00       state changed 10.0.0.18:6000=SORT
1411.15    0.00       state changed 10.0.0.19:6000=SORT
1420.98    0.00       state changed 10.0.0.8:6000=SORT
1487.17    0.00       state changed 10.0.0.17:6000=SHUFFLE
1512.22    0.00       state changed 10.0.0.12:6000=SORT
1677.07    0.00       state changed 10.0.0.16:6000=SHUFFLE
1680.51    0.00       state changed 10.0.0.11:6000=SORT
1723.06    0.00       state changed 10.0.0.18:6000=SHUFFLE
1792.86    0.00       state changed 10.0.0.19:6000=SHUFFLE
1814.83    0.00       state changed 10.0.0.8:6000=SHUFFLE
1927.40    0.00       state changed 10.0.0.12:6000=SHUFFLE
2119.42    0.00       state changed 10.0.0.11:6000=SHUFFLE
2550.57    0.00       state changed 10.0.0.18:6000=MERGE
2550.57    0.00       state changed 10.0.0.18:6000=UPLOAD
2550.60    0.00       state changed 10.0.0.11:6000=MERGE
2550.79    0.00       state changed 10.0.0.15:6000=MERGE
2550.79    0.00       state changed 10.0.0.15:6000=UPLOAD
2550.87    0.00       state changed 10.0.0.19:6000=MERGE
2550.87    0.00       state changed 10.0.0.19:6000=UPLOAD
2550.91    0.00       state changed 10.0.0.12:6000=MERGE
2550.92    0.00       state changed 10.0.0.12:6000=UPLOAD
2550.93    0.00       state changed 10.0.0.16:6000=MERGE
2550.93    0.00       state changed 10.0.0.16:6000=UPLOAD
2551.01    0.00       state changed 10.0.0.17:6000=MERGE
2551.02    0.00       state changed 10.0.0.17:6000=UPLOAD
2551.03    0.00       state changed 10.0.0.14:6000=MERGE
2551.04    0.00       state changed 10.0.0.14:6000=UPLOAD
2551.06    0.00       state changed 10.0.0.8:6000=MERGE
2551.07    0.00       state changed 10.0.0.8:6000=UPLOAD
2551.10    0.00       state changed 10.0.0.13:6000=MERGE
2551.10    0.00       state changed 10.0.0.13:6000=UPLOAD
2559.28    0.00       state changed 10.0.0.11:6000=UPLOAD
3588.45    0.00       state changed 10.0.0.19:6000=DONE
3598.97    0.00       state changed 10.0.0.17:6000=DONE
3607.11    0.00       state changed 10.0.0.13:6000=DONE
3614.02    0.00       state changed 10.0.0.18:6000=DONE
3623.84    0.00       state changed 10.0.0.16:6000=DONE
3636.27    0.00       state changed 10.0.0.12:6000=DONE
3646.11    0.00       state changed 10.0.0.15:6000=DONE
3662.00    0.00       state changed 10.0.0.14:6000=DONE
3694.73    0.00       state changed 10.0.0.11:6000=DONE
3721.79    0.00       state changed 10.0.0.8:6000=DONE
```

### Example run using 1 controller (8 vCPUs, 16GB, 240GB SSD) and 5 workers (4 vCPUs, 8GB, 160GB SSD)

In this setting, each node has to sort 16GB by itself using about 6.5GB of buffer memory and then shuffle its data over the network.

The obvious problem here is that the single controller / data storage caps out on network/disk speed during data fetching.
Can be solved by using an external service like S3 or do multiple instances of minio, which perform load balancing.

```
TIMESTAMP  PROGRESS   MESSAGE
0.00       0.00       Job scheduled
0.01       0.00       Target file has 80 GB of data
0.36       0.00       state changed 10.0.0.4:6000=DOWNLOAD
1.00       0.00       state changed 10.0.0.6:6000=DOWNLOAD
1.65       0.00       state changed 10.0.0.7:6000=DOWNLOAD
2.24       0.00       state changed 10.0.0.5:6000=DOWNLOAD
9.75       0.00       state changed 10.0.0.3:6000=DOWNLOAD
166.78     0.00       state changed 10.0.0.4:6000=SORT
178.92     0.00       state changed 10.0.0.7:6000=SORT
187.43     0.00       state changed 10.0.0.5:6000=SORT
222.27     0.00       state changed 10.0.0.4:6000=SHUFFLE
236.67     0.00       state changed 10.0.0.7:6000=SHUFFLE
243.47     0.00       state changed 10.0.0.5:6000=SHUFFLE
245.19     0.00       state changed 10.0.0.3:6000=SORT
261.11     0.00       state changed 10.0.0.6:6000=SORT
305.01     0.00       state changed 10.0.0.3:6000=SHUFFLE
330.77     0.00       state changed 10.0.0.6:6000=SHUFFLE
396.68     0.00       state changed 10.0.0.6:6000=MERGE
397.12     0.00       state changed 10.0.0.3:6000=MERGE
397.13     0.00       state changed 10.0.0.3:6000=UPLOAD
397.13     0.00       state changed 10.0.0.7:6000=MERGE
397.14     0.00       state changed 10.0.0.7:6000=UPLOAD
397.15     0.00       state changed 10.0.0.5:6000=MERGE
397.15     0.00       state changed 10.0.0.5:6000=UPLOAD
397.18     0.00       state changed 10.0.0.4:6000=MERGE
397.18     0.00       state changed 10.0.0.4:6000=UPLOAD
400.21     0.00       state changed 10.0.0.6:6000=UPLOAD
439.07     0.00       state changed 10.0.0.4:6000=DONE
439.88     0.00       state changed 10.0.0.5:6000=DONE
446.41     0.00       state changed 10.0.0.3:6000=DONE
446.44     0.00       state changed 10.0.0.7:6000=DONE
455.38     0.00       state changed 10.0.0.6:6000=DONE
```

As you can see, the total time it takes to upload/download is roughly 3 minutes.
The sort itself takes about 3 minutes. 

At the moment, sorting performance is largely dominated by the run generation. The sort itself is very expensive due to the size of the entries. More tests may be needed to determine ways to optimize the external sort.