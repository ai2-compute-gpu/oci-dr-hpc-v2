# OCI BM GPU  Shapes
ref: https://docs.oracle.com/en-us/iaas/Content/Compute/References/computeshapes.htm

| **Shape**                 | **OCPU** | **GPU Memory (GB)** | **CPU Memory (GB)** | **Local Disk**              | **Max Network Bandwidth**             | **Max VNICs Total: Linux** | **Max VNICs Total: Windows**      |
|---------------------------|---------:|--------------------:|--------------------:|-----------------------------|---------------------------------------|---------------------------:|-----------------------------------|
| BM.GPU2.2 (2xP100)        |       28 |                  32 |                 192 | Block storage only          | 2 x 25 Gbps                           |                         28 | 15 (1 on first NIC, 14 on second) |
| BM.GPU3.8 (8xV100)        |       52 |                 128 |                 768 | Block storage only          | 2 x 25 Gbps                           |                         52 | 27 (1 on first NIC, 26 on second) |
| BM.GPU4.8 (8xA100)        |       64 |                 320 |                2048 | 27.2 TB NVMe SSD (4 drives) | 1 x 50 Gbps<br>8 x 200 Gbps RDMA      |                         64 | Not supported                     |
| BM.GPU.A10.4 (4xA10)      |       64 |                  96 |                1024 | 7.68 TB NVMe SSD (2 drives) | 2 x 50 Gbps                           |                        256 | Not supported                     |
| BM.GPU.A100-v2.8 (8xA100) |      128 |                 640 |                2048 | 27.2 TB NVMe SSD (4 drives) | 2 x 50 Gbps<br>16 x 100 Gbps RDMA     |                        256 | Not supported                     |
| BM.GPU.MI300X.8           |      112 |                1536 |                2048 | 8 x 3.84 TB NVMe            | 1 x 100 Gbps<br>8 x 400 Gbps RDMA     |                        256 | Not supported                     |
| BM.GPU.L40S.4             |      112 |                 192 |                1024 | 2 x 3.84 TB NVMe            | 1 x 200 Gbps<br>800 Gbps RDMA         |                        256 | Not supported                     |
| BM.GPU.H100.8 (8xH100)    |      112 |                 640 |                2048 | 16 x 3.84 TB NVMe           | 1 x 100 Gbps<br>8 x 2 x 200 Gbps RDMA |                        256 | Not supported                     |
| BM.GPU.H200.8 (8xH200)    |      112 |                1128 |                3072 | 8 x 3.84 TB NVMe            | 1 x 200 Gbps<br>8 x 400 Gbps RDMA     |                        256 | Not supported                     |
| BM.GPU.B200.8 (8xB200)    |      128 |                1440 |                4096 | 8 x 3.84 TB NVMe            | 2 x 200 Gbps<br>8 x 400 Gbps RDMA     |                        256 | Not supported                     |
| BM.GPU.GB200.4 (4xB200)   |      144 |                 768 |                 960 | 4 x 7.68 TB NVMe            | 2 x 200 Gbps<br>4 x 400 Gbps RDMA     |                        512 | Not supported                     |
