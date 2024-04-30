## deployment.yaml
- 这里举例使用了单个文件，如果使用kustomize，则可以替换patch文件的内容
- 如果使用helm部署，则可以不使用这种替换方式，因为helm支持命令传参来设置
```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: #{AppName}
  name: #{AppName}
  namespace: #{Namepspace}
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: #{AppName}
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: #{AppName}
    spec:
      containers:
        - image: #{Image}
          imagePullPolicy: IfNotPresent
          name: app
          resources:
            limits:
              cpu: 1000m
              memory: 1Gi
            requests:
              cpu: 50m
              memory: 50Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      imagePullSecrets:
        - name: tcr-scq
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
```