# Диаграммы сервисов FinFlow

## Диаграмма связи сервисов

```mermaid
graph TB
    subgraph "Клиентский уровень"
        MobileApp[Мобильное приложение]
    end

    subgraph "Инфраструктура"
        PG[(PostgreSQL)]
        Redis[(Redis)]
        MinIO[(MinIO)]
    end

    subgraph "Микросервисы"
        Auth[ff-auth<br/>:8084<br/>Аутентификация]
        ID[ff-id<br/>:8081<br/>Профили пользователей]
        Split[ff-split<br/>:8080<br/>Управление расходами]
        Files[ff-files<br/>:8082<br/>Файловое хранилище]
        TVM[ff-tvm<br/>:8083<br/>Авторизация сервисов]
    end

    %% Взаимодействие клиента с сервисами
    MobileApp --> Auth
    MobileApp --> ID
    MobileApp --> Split
    MobileApp --> Files

    %% Межсервисное взаимодействие
    Auth --> ID
    Split --> ID
    Split --> Files
    ID --> Files

    %% TVM авторизация
    Auth --> TVM
    ID --> TVM
    Split --> TVM
    Files --> TVM

    %% Подключения к БД
    Auth --> PG
    ID --> PG
    Split --> PG
    TVM --> PG
    TVM --> Redis

    %% Файловое хранилище
    Files --> MinIO

    %% Стили
    classDef service fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef database fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef client fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px

    class Auth,ID,Split,Files,TVM service
    class PG,Redis,MinIO database
    class MobileApp client
```

## Диаграммы классов

### ff-auth - Сервис аутентификации

```mermaid
classDiagram
    class User {
        +int64 ID
        +string Email
        +string PasswordHash
        +string Nickname
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +[]UserRole Roles
        +[]Session Sessions
        +[]LoginHistory LoginHistory
        +[]Device Devices
        +BeforeUpdate() error
    }

    class RoleEntity {
        +int ID
        +string Name
    }

    class UserRole {
        +int64 UserID
        +int RoleID
    }

    class Session {
        +uuid.UUID ID
        +int64 UserID
        +string RefreshToken
        +[]string IPAddress
        +time.Time ExpiresAt
        +time.Time CreatedAt
    }

    class LoginHistory {
        +int ID
        +int64 UserID
        +string IPAddress
        +string UserAgent
        +time.Time CreatedAt
    }

    class Device {
        +int ID
        +int64 UserID
        +string DeviceID
        +string UserAgent
        +time.Time LastLogin
    }

    class KeyPair {
        +int ID
        +string PublicKey
        +string PrivateKey
        +bool IsActive
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +BeforeUpdate() error
    }

    User --> UserRole : has
    RoleEntity --> UserRole : belongs_to
    User --> Session : has
    User --> LoginHistory : has
    User --> Device : uses
```

### ff-id - Сервис профилей пользователей

```mermaid
classDiagram
    class User {
        +int64 ID
        +string Email
        +string Phone
        +string Nickname
        +string Name
        +time.Time Birthdate
        +uuid.UUID AvatarID
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +[]UserAvatar Avatars
        +[]UserFriend Friends
        +BeforeUpdate() error
        +AfterFind() error
    }

    class UserAvatar {
        +uuid.UUID ID
        +int64 UserID
        +uuid.UUID FileID
        +time.Time UploadedAt
        +AfterFind() error
    }

    class UserFriend {
        +int64 ID
        +int64 UserID
        +int64 FriendID
        +string Status
        +time.Time CreatedAt
        +User User
        +User Friend
        +AfterFind() error
    }

    User --> UserAvatar : has
    User --> UserFriend : has
    UserFriend --> User : friend
```

### ff-split - Сервис управления расходами

```mermaid
classDiagram
    class User {
        +int64 ID
        +int64 UserID
        +string NicknameCashed
        +string NameCashed
        +string PhotoUUIDCashed
        +bool IsDummy
        +[]Event Events
        +[]Activity Activities
        +[]Transaction Transactions
        +[]Task Tasks
        +[]TransactionShare Shares
        +[]Debt DebtsFrom
        +[]Debt DebtsTo
    }

    class Event {
        +int64 ID
        +string Name
        +string Description
        +int CategoryID
        +string ImageID
        +string Status
        +EventCategory Category
        +[]User Users
        +[]Activity Activities
        +[]Transaction Transactions
        +[]Task Tasks
    }

    class EventCategory {
        +int ID
        +string Name
        +int IconID
        +Icon Icon
        +[]Event Events
    }

    class Transaction {
        +int ID
        +int64 EventID
        +string Name
        +int TransactionCategoryID
        +time.Time Datetime
        +float64 TotalPaid
        +int64 PayerID
        +int SplitType
        +Event Event
        +TransactionCategory TransactionCategory
        +User Payer
        +[]TransactionShare Shares
        +[]Debt Debts
    }

    class TransactionShare {
        +int ID
        +int TransactionID
        +int64 UserID
        +float64 Value
        +Transaction Transaction
        +User User
    }

    class Debt {
        +int ID
        +int TransactionID
        +int64 FromUserID
        +int64 ToUserID
        +float64 Amount
        +Transaction Transaction
        +User FromUser
        +User ToUser
    }

    class OptimizedDebt {
        +int ID
        +int64 EventID
        +int64 FromUserID
        +int64 ToUserID
        +float64 Amount
        +time.Time CreatedAt
        +time.Time UpdatedAt
    }

    class TransactionCategory {
        +int ID
        +string Name
        +int IconID
        +Icon Icon
        +[]Transaction Transactions
    }

    class Icon {
        +int ID
        +string Name
        +string FileUUID
        +[]TransactionCategory TransactionCategories
    }

    class Activity {
        +int ID
        +int64 EventID
        +int64 UserID
        +string Description
        +int IconID
        +time.Time CreatedAt
        +Icon Icon
        +Event Event
        +User User
    }

    class Task {
        +int ID
        +int64 UserID
        +int64 EventID
        +string Title
        +string Description
        +int Priority
        +time.Time CreatedAt
        +User User
        +Event Event
    }

    Event --> Transaction : has
    Event --> Activity : has
    Event --> Task : has
    Event --> EventCategory : belongs_to
    Event --> User : participants

    Transaction --> TransactionShare : split_into
    Transaction --> Debt : creates
    Transaction --> User : paid_by
    Transaction --> TransactionCategory : categorized

    TransactionShare --> User : belongs_to
    Debt --> User : from
    Debt --> User : to

    Activity --> User : performed_by
    Activity --> Icon : has
    Task --> User : assigned_to

    TransactionCategory --> Icon : has
    EventCategory --> Icon : has
```

### ff-files - Сервис файлового хранилища

```mermaid
classDiagram
    class File {
        +uuid.UUID ID
        +string OriginalName
        +string ContentType
        +int64 Size
        +string Path
        +string Hash
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +[]FileAccess Access
    }

    class FileAccess {
        +int ID
        +uuid.UUID FileID
        +int64 UserID
        +string AccessType
        +time.Time GrantedAt
        +time.Time ExpiresAt
        +File File
    }

    class UploadSession {
        +uuid.UUID ID
        +int64 UserID
        +string Status
        +int64 TotalSize
        +int64 UploadedSize
        +time.Time CreatedAt
        +time.Time ExpiresAt
    }

    File --> FileAccess : has
    FileAccess --> File : belongs_to
```

### ff-tvm - Сервис авторизации сервисов

```mermaid
classDiagram
    class Service {
        +int ID
        +string Name
        +string PublicKey
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +[]KeyPair KeyPairs
        +[]ServiceAccess AccessFrom
        +[]ServiceAccess AccessTo
    }

    class ServiceAccess {
        +int ID
        +int FromID
        +int ToID
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +Service FromService
        +Service ToService
    }

    class KeyPair {
        +int ID
        +int ServiceID
        +string PublicKey
        +string PrivateKey
        +bool IsActive
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +Service Service
    }

    class TicketPayload {
        +int From
        +int To
        +int64 TTL
        +string Metadata
        +time.Time IssuedAt
    }

    class Ticket {
        +TicketPayload Payload
        +string Signature
        +Validate() bool
        +IsExpired() bool
    }

    Service --> KeyPair : has
    Service --> ServiceAccess : from
    Service --> ServiceAccess : to
    ServiceAccess --> Service : from
    ServiceAccess --> Service : to
```

## Архитектурные особенности

### Слоистая архитектура сервисов

```mermaid
graph TB
    subgraph "Каждый микросервис"
        API[API Layer<br/>REST Controllers]
        BL[Business Logic<br/>Services]
        REPO[Repository Layer<br/>Data Access]
        DB[(Database)]
    end

    API --> BL
    BL --> REPO
    REPO --> DB
```

### Модель безопасности

```mermaid
graph LR
    Client[Клиент] -->|JWT Token| API[API Gateway]
    API -->|Валидация| Auth[ff-auth]
    
    subgraph "Межсервисное взаимодействие"
        Service1[Сервис A] -->|TVM Request| TVM[ff-tvm]
        TVM -->|TVM Ticket| Service1
        Service1 -->|Request + TVM Ticket| Service2[Сервис B]
        Service2 -->|Validation| TVM
    end
```

## Основные компоненты системы

### Уровни архитектуры

1. **Клиентский уровень**: Мобильное приложение
2. **Микросервисы**: Независимые сервисы с собственной БД
3. **Инфраструктура**: PostgreSQL, Redis, MinIO

### Принципы проектирования

- **Микросервисная архитектура**: Каждый сервис отвечает за свою доменную область
- **Database per Service**: У каждого сервиса своя база данных
- **API-First**: Все взаимодействие через REST API
- **Независимое развертывание**: Сервисы могут обновляться независимо
- **Отказоустойчивость**: Graceful degradation при недоступности сервисов

### Безопасность

- **JWT токены** для аутентификации пользователей
- **TVM тикеты** для межсервисной авторизации
- **ED25519 подписи** для криптографической защиты
- **HTTPS** для всех внешних соединений
- **Принцип минимальных привилегий** для доступа к ресурсам
