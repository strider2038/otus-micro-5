# Otus Microservice architecture Homework 6

## Домашнее задание выполнено для курса ["Microservice architecture"](https://otus.ru/lessons/microservice-architecture/)

### Теоретическая часть

[Исследование вариантов реализации системы](docs/design.md).

Для реализации выбран гибридный вариант (REST + Event collaboration). Для упрощения сервисы объединяют в себе
HTTP серверы и Kafka Consumer'ы, поэтому масштабирование подов ограничено количеством партиций топиков в Kafka.

### TODO

* [x] identity service (dispatch events)
* [x] billing service
  * [x] billing api
  * [x] billing identity consumer
  * [x] billing worker
* [x] order service
* [x] notification service
* [ ] postman tests
* [ ] transactions ?

### Запуск приложения

```shell
# запуск minikube
# версия k8s v1.19, на более поздних есть проблемы с установкой ambassador
minikube start --cpus=6 --memory=16g --disk-size='40000mb' --vm-driver=virtualbox --kubernetes-version="v1.19.0"

kubectl create namespace otus
kubectl config set-context --current --namespace=otus

# установка Ambassador
helm install aes datawire/ambassador -f deploy/ambassador-values.yaml

# установка Apache Kafka
helm install kafka bitnami/kafka -f deploy/kafka-values.yaml

## запуск проекта
helm install --wait -f deploy/identity-values.yaml identity-service ./services/identity-service/deployments/identity-service --atomic

# применение настроек ambassador
kubectl apply -f services/ambassador/
```

### Тестирование

Тесты Postman расположены в директории `test/postman`. Запуск тестов.

```bash
newman run ./test/postman/test.postman_collection.json
```

Или с использованием Docker.

```bash
docker run -v $PWD/test/postman/:/etc/newman --network host -t postman/newman:alpine run test.postman_collection.json
```

### Отладка

```shell
# отладка kafka consumer
kubectl run kafka-client --restart='Never' --image docker.io/bitnami/kafka:2.8.1-debian-10-r73 --namespace otus --command -- sleep infinity
kubectl exec --tty -i kafka-client --namespace otus -- bash
kafka-console-consumer.sh \
    --bootstrap-server kafka.otus.svc.cluster.local:9092 \
    --topic <TOPIC_NAME> \
    --property print.headers=true \
    --from-beginning
    
# отладка postgres
kubectl port-forward svc/identity-service-postgresql 5432:5432
```
