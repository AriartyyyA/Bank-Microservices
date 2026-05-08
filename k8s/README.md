# GoBank в Kubernetes (minikube)

Этот документ описывает полный локальный запуск проекта в Kubernetes:

- поднимаем кластер `minikube`;
- собираем образы сервисов в локальный Docker-демон minikube;
- применяем манифесты `base`, `infra`, `apps`;
- проверяем API, Kafka, Jaeger, Prometheus, Grafana;
- (опционально) включаем Ingress и используем `gobank.local`.

## 1. Что нужно установить заранее

- `kubectl` (CLI для Kubernetes)
- `minikube` (локальный Kubernetes-кластер)
- Docker Desktop/Engine (для сборки образов)
- `migrate` (если применяешь SQL-миграции вручную)

Проверка:

```powershell
kubectl version --client
minikube version
docker version
```

## 2. Запуск minikube-кластера

```powershell
minikube start --cpus=4 --memory=8192
kubectl config current-context
kubectl get nodes
```

Ожидаемо:

- текущий контекст: `minikube`
- нода `minikube` в статусе `Ready`

## 3. Сборка образов для сервисов GoBank

Важно: Kubernetes в minikube должен видеть образы локально.  
Для этого переключаемся на Docker-демон minikube:

```powershell
minikube -p minikube docker-env --shell powershell | Invoke-Expression
```

Дальше собираем 3 сервиса (Dockerfile использует `ARG SERVICE`):

```powershell
docker build --build-arg SERVICE=auth -t gobank/auth:latest .
docker build --build-arg SERVICE=wallet -t gobank/wallet:latest .
docker build --build-arg SERVICE=notification -t gobank/notification:latest .
```

Проверка:

```powershell
docker images | findstr gobank
```

## 4. Применение Kubernetes-манифестов (этап 1, без Ingress)

### 4.1 База: namespace + конфиги

```powershell
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/secret-app.yaml -f k8s/base/configmap-app.yaml
```

### 4.2 Инфраструктура

```powershell
kubectl apply -f k8s/infra/postgres-auth.yaml
kubectl apply -f k8s/infra/postgres-wallet.yaml
kubectl apply -f k8s/infra/redis.yaml
kubectl apply -f k8s/infra/zookeeper.yaml
kubectl apply -f k8s/infra/kafka.yaml
kubectl apply -f k8s/infra/jaeger.yaml
kubectl apply -f k8s/infra/prometheus.yaml
kubectl apply -f k8s/infra/grafana.yaml
```

### 4.3 Приложения

```powershell
kubectl apply -f k8s/apps/auth.yaml
kubectl apply -f k8s/apps/wallet.yaml
kubectl apply -f k8s/apps/notification.yaml
```

## 5. Проверка статуса после деплоя

```powershell
kubectl get pods -n gobank
kubectl get svc -n gobank
```

Если какой-то pod не готов:

```powershell
kubectl describe pod <pod-name> -n gobank
kubectl logs <pod-name> -n gobank --tail=200
```

Если deployment долго "висит":

```powershell
kubectl rollout status deployment/<name> -n gobank --timeout=60s
```

## 6. Миграции БД в Kubernetes

Даже если миграции применялись в `docker-compose`, в k8s это отдельные базы.  
Нужно применить их для `postgres-auth` и `postgres-wallet`.

Открой два терминала для port-forward БД:

```powershell
kubectl port-forward -n gobank svc/postgres-auth 5432:5432
kubectl port-forward -n gobank svc/postgres-wallet 5433:5432
```

В третьем терминале:

```powershell
make migrate_auth
make migrate_wallet
```

Или напрямую:

```powershell
migrate -path migrations/auth -database "postgres://gobank:secret@localhost:5432/gobank_auth?sslmode=disable" up
migrate -path migrations/wallet -database "postgres://gobank:secret@localhost:5433/gobank_wallet?sslmode=disable" up
```

## 7. Доступ к сервисам через port-forward

Открой отдельные терминалы и держи команды запущенными:

```powershell
kubectl port-forward -n gobank svc/auth 8080:8080
kubectl port-forward -n gobank svc/wallet 8081:8080
kubectl port-forward -n gobank svc/jaeger 16686:16686
kubectl port-forward -n gobank svc/prometheus 9091:9090
kubectl port-forward -n gobank svc/grafana 3000:3000
```

Если `localhost:8081` не доступен после rollout/restart, обычно просто отвалился `port-forward` — запусти его заново.

## 8. Чеклист валидации (обязательно)

### 8.1 Health endpoints

- `http://localhost:8080/healthz`
- `http://localhost:8081/healthz`

Оба должны вернуть `200 OK`.

### 8.2 API сценарий

1. Регистрация и login в `auth`
2. Создание кошельков
3. `POST /wallets/transfer`

### 8.3 Kafka цепочка

Проверь, что `notification` читает событие:

```powershell
kubectl logs -n gobank deploy/notification -f
```

### 8.4 Tracing (Jaeger)

- открыть `http://localhost:16686`
- найти trace сервиса `wallet-service`
- убедиться, что в trace есть `auth-service` и `notification-service`
- убедиться, что видны бизнес-атрибуты (`transfer.transaction_id`, `transfer.amount` и т.д.)

### 8.5 Metrics

- открыть `http://localhost:9091/targets`
- убедиться, что `gobank-auth` и `gobank-wallet` в состоянии `UP`

### 8.6 Grafana

- открыть `http://localhost:3000`
- login: `admin` / `admin`
- datasource Prometheus:
  - URL: `http://prometheus.gobank.svc.cluster.local:9090`

## 9. Ingress (этап 2)

### 9.1 Включить ingress-контроллер

```powershell
minikube addons enable ingress
kubectl apply -f k8s/ingress/api-ingress.yaml
```

### 9.2 Добавить запись в hosts

Узнать IP minikube:

```powershell
minikube ip
```

Добавить в hosts:

```text
<minikube-ip> gobank.local
```

### 9.3 Проверка маршрутов

- `http://gobank.local/auth/swagger/index.html`
- `http://gobank.local/wallet/swagger/index.html`

## 10. Частые проблемы и быстрые решения

### Проблема: `relation "users" does not exist`
Причина: не применены миграции в k8s-базе.  
Решение: выполнить раздел "Миграции БД в Kubernetes".

### Проблема: `ECONNREFUSED 127.0.0.1:8081` в Postman
Причина: не запущен `kubectl port-forward` для `wallet`.  
Решение: снова выполнить `kubectl port-forward -n gobank svc/wallet 8081:8080`.

### Проблема: `notification` не появляется в trace
Причина: `notification` стартовал раньше Kafka и завершил consumer-loop.  
Решение:

```powershell
kubectl rollout restart deployment/notification -n gobank
```

### Проблема: `kubectl rollout status` "висит"
Причина: rollout не завершился, есть плохая ревизия.  
Решение: использовать таймаут и смотреть историю rollout:

```powershell
kubectl rollout status deployment/<name> -n gobank --timeout=30s
kubectl rollout history deployment/<name> -n gobank
```

## 11. Очистка окружения

Удалить весь namespace проекта:

```powershell
kubectl delete namespace gobank
```
