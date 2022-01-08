# Варианты реализации системы заказов

[Спецификация публичного REST API](public-api.yaml)

## 1. Взаимодействие через HTTP

### Спецификации

* [Спецификация REST API сервисов](1_http/rest-api.yaml)

### Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/identity/register
    API Gateway->>Identity Service: POST /api/v1/register
    Identity Service->>Billing Service: PUT /api/v1/accounts/{userId}
    Billing Service-->>Identity Service: 201 Created (account)
    Identity Service-->>API Gateway: 201 Created (user)
    API Gateway-->>Client: 201 Created (user)
```

### Создание заказа

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/ordering/orders
    API Gateway->>Order Service: POST /api/v1/orders
    Order Service->>Billing Service: POST /api/v1/payments
    
    alt Order succeded
        Billing Service-->>Order Service: 200 OK
        Order Service->>Notification Service: POST /api/v1/notifications (success)
        Notification Service->>Identity Service: GET /api/v1/users/{id}
        Identity Service-->>Notification Service: 200 OK (user)
        Notification Service->>Client: Order succeded email
        Notification Service-->>Order Service: 200 OK
        Order Service-->>API Gateway: 200 OK
        API Gateway-->>Client: 200 OK
    else Order failed
        Billing Service-->>Order Service: 422 Unprocessable entity (not enough money)
        Order Service->>Notification Service: POST /api/v1/notifications (fail)
        Notification Service->>Identity Service: GET /api/v1/users/{id}
        Identity Service-->>Notification Service: 200 OK (user)
        Notification Service->>Client: Order failed email
        Notification Service-->>Order Service: 200 OK
        Order Service-->>API Gateway: 422 Unprocessable entity (not enough money)
        API Gateway-->>Client: 422 Unprocessable entity (not enough money)
    end
```

### Выводы

Достоинства:

* простота реализации;
* пользователь узнает о результате операции сразу (синхронное взаимодействие).

Недостатки:

* большое время ответа из-за синхронного взаимодействия между сервисами;
* асинхронные операции (отправка уведомлений) выполняются синхронно;
* необходим механизм оркестрации из-за распределенности обработки.

## 2. Событийные взаимодействия

### Спецификации

* [Спецификация REST API сервисов](2_events/rest-api.yaml)
* [Спецификация Async API сервисов](2_events/async-api.yaml)

### Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/identity/register
    API Gateway->>Message broker: publish RegistrationRequested
    Message broker-->>Identity Service: consume RegistrationRequested
    Identity Service->>Identity Service: create user
    Identity Service->>Message broker: publish UserCreated 
    
    par Complete registration
        Message broker-->>API Gateway: consume UserCreated
        API Gateway->>Identity Service: GET /api/v1/users/{userId}
        Identity Service-->>API Gateway: 200 OK (user)
        API Gateway-->>Client: 201 Created
    and Create billing account
        Message broker-->>Billing Service: consume UserCreated
        Billing Service->>Billing Service: create user account
    end
```

### Создание заказа

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/ordering/orders
    API Gateway->>Message broker: publish CreateOrder
    Message broker-->>Order Service: consume CreateOrder
    Order Service->>Order Service: Create order with status pending
    Order Service->>Message broker: publish PaymentRequired
    Message broker-->>Billing Service: consume PaymentRequired
    Billing Service->>Order Service: GET /api/v1/orders/{id}
    Order Service-->>Billing Service: 200 OK (order)
    
    alt Order succeded
        Billing Service->>Message broker: publish PaymentCreated
        Message broker-->>Order Service: consume PaymentCreated
        Order Service->>Order Service: Update order with status succeded
        Order Service->>Message broker: publish OrderCreated
        par Send response to client
            Message broker-->>API Gateway: consume OrderCreated
            API Gateway-->>Client: 200 OK
        and Send notification to client
            Message broker-->>Notification Service: consume OrderCreated
            Notification Service->>Order Service: GET /api/v1/orders/{id}
            Order Service-->>Notification Service: 200 OK (order)
            Notification Service->>Identity Service: GET /api/v1/users/{id}
            Identity Service-->>Notification Service: 200 OK (user)
            Notification Service->>Client: Order succeded email
        end
    else Order failed
        Billing Service->>Message broker: publish PaymentFailed
        Message broker-->>Order Service: consume PaymentFailed
        Order Service->>Order Service: Update order with status failed
        Order Service->>Message broker: publish OrderFailed
        par Send response to client
            Message broker-->>API Gateway: consume OrderFailed
            API Gateway-->>Client: 422 Unprocessable entity (not enough money)
        and Send notification to client
            Message broker-->>Notification Service: consume OrderFailed
            Notification Service->>Order Service: GET /api/v1/orders/{id}
            Order Service-->>Notification Service: 200 OK (order)
            Notification Service->>Identity Service: GET /api/v1/users/{id}
            Identity Service-->>Notification Service: 200 OK (user)
            Notification Service->>Client: Order failed email
        end
    end
```

### Выводы

Достоинства:

* service discovery на основе брокера сообщений => легче масштабирование;
* создание аккаунта в биллинге осуществляется асинхронно;
* отправка уведомлений осуществляется асинхронно (параллельно с ответом в API Gateway).

Недостатки:

* повышенная сложность реализации: 
  * логика работы с брокером в API Gateway;
  * присутствуют синхронные запросы между сервисами;
  * API Gateway должен хранить состояние (correlationId).
* дополнительная связность между сервисами (биллинг обращается к заказам);
* увеличенное время ответа из-за дополнительного взаимодействия через брокер сообщений;
* проблемы асинхронного создания аккаунта в биллинге:
  * может возникнуть проблема отсутствия аккаунта в случае сбоя;
  * критично время выполнения события (аккаунта еще может не быть, когда он уже может быть нужен - маловероятно, но возможно).

## 3. Event collaboration

### Спецификации

* [Спецификация REST API сервисов](3_event-collaboration/rest-api.yaml)
* [Спецификация Async API сервисов](3_event-collaboration/async-api.yaml)

### Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/identity/register
    API Gateway->>Message broker: publish RegistrationRequested
    Message broker-->>Identity Service: consume RegistrationRequested
    Identity Service->>Identity Service: create user
    Identity Service->>Message broker: publish UserCreated 
    
    par Complete registration
        Message broker-->>API Gateway: consume UserCreated
        API Gateway-->>Client: 201 Created
    and Create billing account
        Message broker-->>Billing Service: consume UserCreated
        Billing Service->>Billing Service: create billing account
    and Create notification profile
        Message broker-->>Notification Service: consume UserCreated
        Notification Service->>Notification Service: create notification profile
    end
```

### Создание заказа

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/ordering/orders
    API Gateway->>Message broker: publish CreateOrder
    Message broker-->>Order Service: consume CreateOrder
    Order Service->>Order Service: Create order with status pending
    Order Service->>Message broker: publish CreatePayment
    Message broker-->>Billing Service: consume CreatePayment
    
    alt Order succeded
        Billing Service->>Message broker: publish PaymentCreated
        Message broker-->>Order Service: consume PaymentCreated
        Order Service->>Order Service: Update order with status succeded
        Order Service->>Message broker: publish OrderCreated
        par Send response to client
            Message broker-->>API Gateway: consume OrderCreated
            API Gateway-->>Client: 200 OK
        and Send notification to client
            Message broker-->>Notification Service: consume OrderCreated
            Notification Service->>Client: Order succeded email
        end
    else Order failed
        Billing Service->>Message broker: publish PaymentFailed
        Message broker-->>Order Service: consume PaymentFailed
        Order Service->>Order Service: Update order with status failed
        Order Service->>Message broker: publish OrderFailed
        par Send response to client
            Message broker-->>API Gateway: consume OrderFailed
            API Gateway-->>Client: 422 Unprocessable entity (not enough money)
        and Send notification to client
            Message broker-->>Notification Service: consume OrderFailed
            Notification Service->>Client: Order failed email
        end
    end
```

### Выводы

Достоинства:

* service discovery на основе брокера сообщений => легче масштабирование;
* нет API вызовов между сервисами;
* создание аккаунта в биллинге осуществляется асинхронно;
* отправка уведомлений осуществляется асинхронно (параллельно с ответом в API Gateway).
* низкая связность между сервисами (биллинг ничего не знает о сервисе заказов);

Недостатки:

* повышенная сложность реализации:
  * логика работы с брокером в API Gateway;
  * API Gateway должен хранить состояние (correlationId).
* проблемы асинхронного создания аккаунта в биллинге:
  * может возникнуть проблема отсутствия аккаунта в случае сбоя;
  * критично время выполнения события (аккаунта еще может не быть, когда он уже может быть нужен - маловероятно, но возможно).
* сервисы должны хранить состояние из других областей (например, сервис уведомлений хранит копию сведений о пользователе).

## 4. Гибридный вариант

### Спецификации

* [Спецификация REST API сервисов](4_hybrid/rest-api.yaml)
* [Спецификация Async API сервисов](4_hybrid/async-api.yaml)

### Регистрация пользователя

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/identity/register
    API Gateway->>Identity Service: POST /api/v1/register
    Identity Service->>Identity Service: create user
    Identity Service->>Message broker: publish UserCreated
    
    par Complete registration
        Identity Service-->>API Gateway: 201 Created (user)
        API Gateway-->>Client: 201 Created (user)
    and Create billing account
        Message broker-->>Billing Service: consume UserCreated
        Billing Service->>Billing Service: create billing account
    and Create notification profile
        Message broker-->>Notification Service: consume UserCreated
        Notification Service->>Notification Service: create notification profile
    end
```

### Создание заказа

```mermaid
sequenceDiagram
    actor Client
    
    Client->>API Gateway: POST /api/v1/ordering/orders
    API Gateway->>Order Service: POST /api/v1/orders
    Order Service->>Order Service: Create order with status pending
    Order Service-->>API Gateway: 202 Accepted
    API Gateway-->>Client: 202 Accepted
    Order Service->>Message broker: publish CreatePayment
    Message broker-->>Billing Service: consume CreatePayment
    
    alt Order succeded
        Billing Service->>Message broker: publish PaymentCreated
        Message broker-->>Order Service: consume PaymentCreated
        Order Service->>Order Service: Update order with status succeded
        Order Service->>Message broker: publish OrderCreated
        par Order notification
            Message broker-->>Notification Service: consume OrderCreated
            Notification Service->>Client: Order succeded email
        and Order push event
            Message broker-->>Push Service: consume OrderCreated
            Push Service->>Client: HTTP/2 push event
        end
    else Order failed
        Billing Service->>Message broker: publish PaymentFailed
        Message broker-->>Order Service: consume PaymentFailed
        Order Service->>Order Service: Update order with status failed
        Order Service->>Message broker: publish OrderFailed
        par Order notification
            Message broker-->>Notification Service: consume OrderFailed
            Notification Service->>Client: Order failed email
        and Order push event
            Message broker-->>Push Service: consume OrderFailed
            Push Service->>Client: HTTP/2 push event
        end
    end
```

### Выводы

Достоинства

* баланс между производительностью и простотой:
  * API Gateway выполняет роль proxy и не работает с брокером сообщений;
  * ответственность за асинхронность операций лежит на самих сервисах:
    * простые операции можно делать в синхронном стиле (пополнение/снятие средств с аккаунта в биллинге);
    * сложные операции в асинхронном (регистрация, оформление заказа).
* API создания заказа работает асинхронно с клиентом. Ожидание операции можно реализовать двумя путями:
  * через поллинг GET /api/v1/orders/{orderId};
  * через HTTP/2 push event или веб-сокеты.
* отсутствует необходимость хранения состояния (correlationId);
* низкая связанность между сервисами:
  * платежи ничего не знают о том, кто их инициировал (сервис заказов или другой);
  * платеж может создать любой сервис, а не только сервис заказов.

Недостатки:

* проблемы асинхронного создания аккаунта в биллинге и сервисе уведомлений:
  * может возникнуть проблема отсутствия аккаунта в случае сбоя;
  * критично время выполнения события (аккаунта еще может не быть, когда он уже может быть нужен - маловероятно, но возможно).
* сервисы должны хранить состояние из других областей (например, сервис уведомлений хранит копию сведений о пользователе).
