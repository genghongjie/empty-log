中文
# 清理K8S应用日志

清理方式只将日志文件置空，不会删除日志文件

## 清理方式

1 设置总日志文件夹大小阈值
   计算文件夹总大小，一旦超过阈值，将会把所有日志目录下所有文件的内容清空

2 设置单个文件的大小阈值
   遍历日志目录下所有文件，如日志文件大小超过阈值，则将此日志文件内容清空

说明：

容器内扫描日志目录为 /data/logs

请将此容器目录挂载到本地日志目录

环境变量说明：

- folder_max_size 日志总文件夹大小阈值  默认值 10G
- file_max_size      单个日志文件内容大小阈值 默认值 2G
- cron                     定时规则     默认 0 * */1 * * ?  每隔一个小时执行一次

英文
# clean up K8S app log

Cleanup only leaves the log file empty, not deleted

## cleaning method

Set the total log folder size threshold
Calculate the total folder size, and once the threshold is checked, the contents of all files in all log directories are cleared

Set the size threshold for a single file
Traverse all files in the log directory and clear the log file contents if the log file size exceeds the threshold

Description:

The scan log directory in the container is /data/logs
Mount this container directory to the local log directory

Description of environment variables:

- folder_max_size log total folder size threshold default value 10G
- Single log file content size threshold defaults to 2G
- cron     Timing rule defaults to 0 * */1 * *? Perform every hour


k8s yaml
```
apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "15"
    field.cattle.io/creatorId: user-trs48
  creationTimestamp: "2019-11-26T02:52:40Z"
  generation: 15
  labels:
    cattle.io/creator: norman
    workload.user.cattle.io/workloadselector: daemonSet-clean-logs-empty-log-file
  name: empty-log
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      workload.user.cattle.io/workloadselector: daemonSet-clean-logs-empty-log-file
  template:
    metadata:
      annotations:
        cattle.io/timestamp: "2019-11-26T08:06:12Z"
      creationTimestamp: null
      labels:
        workload.user.cattle.io/workloadselector: daemonSet-clean-logs-empty-log-file
    spec:
      containers:
      - env:
        - name: cron
          value: 0 * */1 * * ?
        - name: file_max_size
          value: "1"
        - name: folder_max_size
          value: "10"
        image: yinjianxia/empty-log:0.1
        imagePullPolicy: Always
        name: empty-log
        resources: {}
        securityContext:
          allowPrivilegeEscalation: false
          capabilities: {}
          privileged: false
          readOnlyRootFilesystem: false
          runAsNonRoot: false
        stdin: true
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        tty: true
        volumeMounts:
        - mountPath: /data/logs
          name: vol1
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
```