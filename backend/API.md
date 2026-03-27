# Pixelcraft API Documentation

Local Base URL: `http://localhost:8080/api/v1`
Official Base URL: `https://api.pixelcraft-studio.store/api/v1`

## Authentication

### Register
**POST** `/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "full_name": "John Doe",
  "username": "johndoe",
  "discord_handle": "john#1234",
  "whatsapp_phone": "+5511999999999",
  "cpf": "12345678900",
  "referral_code": "OPTIONAL_CODE"
}
```

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn...",
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "full_name": "John Doe",
    "balance": 0,
    "referral_code": "MYREFCODE",
    "is_admin": false,
    "created_at": "2025-11-27T10:00:00Z",
    "updated_at": "2025-11-27T10:00:00Z"
  }
}
```

### Login
**POST** `/auth/login`

Authenticate a user and retrieve a JWT token.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsIn...",
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "full_name": "John Doe",
    "balance": 10000,
    "referral_code": "MYREFCODE",
    "is_admin": true,
    "created_at": "2025-10-20T00:32:55Z",
    "updated_at": "2025-11-22T00:28:57Z"
  }
}
```

---

## Users

### Get Profile
**GET** `/users/me`

Get the authenticated user's profile.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "id": "uuid-string",
  "email": "user@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "discord_handle": "john#1234",
  "whatsapp_phone": "+5511999999999",
  "balance": 150.00,
  "referral_code": "MYREFCODE",
  "is_admin": false,
  "created_at": "2025-11-27T10:00:00Z",
  "updated_at": "2025-11-27T10:00:00Z"
}
```

### Update Profile
**PUT** `/users/me`

Update the authenticated user's profile.

**Headers:**
`Authorization: Bearer <token>`

**Request Body:**
```json
{
  "full_name": "John Doe Updated",
  "discord_handle": "john_new#1234"
}
```

**Response (200 OK):**
```json
{
  "message": "Profile updated successfully"
}
```

---

## Products

### List Products
**GET** `/products`

Get a paginated list of products.

**Query Parameters:**
- `page`: Page number (default 1)
- `page_size`: Items per page (default 20)
- `type`: Filter by type (PLUGIN, MOD, MAP, TEXTUREPACK, SERVER_TEMPLATE)

**Response (200 OK):**
```json
{
  "products": [
    {
      "id": "uuid-string",
      "name": "Super Plugin",
      "description": "A great plugin",
      "price": 29.99,
      "type": "PLUGIN",
      "is_exclusive": false,
      "image_url": "https://example.com/image.jpg",
      "is_active": true,
      "created_at": "2025-11-27T10:00:00Z",
      "updated_at": "2025-11-27T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

### Get Product
**GET** `/products/:id`

Get details of a specific product.

**Response (200 OK):**
```json
{
  "id": "uuid-string",
  "name": "Super Plugin",
  "description": "A great plugin",
  "price": 29.99,
  "type": "PLUGIN",
  "is_exclusive": false,
  "stock_quantity": 100,
  "image_url": "https://example.com/image.jpg",
  "is_active": true,
  "created_at": "2025-11-27T10:00:00Z",
  "updated_at": "2025-11-27T10:00:00Z"
}
```

### Create Product (Admin)
**POST** `/products`

**Headers:**
`Authorization: Bearer <token>`

**Request Body:**
```json
{
  "name": "New Map",
  "description": "Awesome map",
  "price": 15.00,
  "type": "MAP",
  "download_url": "https://s3.bucket/file.zip",
  "is_exclusive": true,
  "stock_quantity": 10,
  "image_url": "https://example.com/map.jpg"
}
```

**Response (201 Created):**
Returns the created product object.

---

## Checkout

### Process Checkout
**POST** `/checkout`

Process a purchase (products or plans).

**Headers:**
`Authorization: Bearer <token>`

**Request Body:**
```json
{
  "cart": [
    {
      "product_id": "uuid-string-1",
      "quantity": 1
    }
  ],
  "coupon_code": "SUMMER2025",
  "referral_code": "FRIEND123",
  "use_balance": true
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "payment_id": "uuid-string",
  "final_amount": 10.00,
  "discount_applied": 5.00,
  "message": "Purchase successful"
}
```

### Validate Discount
**POST** `/discounts/validate`

Check if a coupon or referral code is valid.

**Headers:**
`Authorization: Bearer <token>`

**Request Body:**
```json
{
  "code": "SUMMER2025",
  "amount": 100.00
}
```

**Response (200 OK):**
```json
{
  "is_valid": true,
  "discount_amount": 10.00,
  "final_amount": 90.00,
  "message": "Discount applied successfully"
}
```

---

## Dashboard

### Get Dashboard Stats
**GET** `/dashboard/stats`

Get user statistics for the dashboard.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "balance": 150.00,
  "total_spent": 500.00,
  "products_purchased": 5,
  "active_subscriptions": 1,
  "recent_payments": [
    {
      "id": "uuid-string",
      "description": "Purchase of Super Plugin",
      "amount": 29.99,
      "status": "COMPLETED",
      "created_at": "2025-11-27T10:00:00Z"
    }
  ],
  "monthly_spending": [
    {
      "month": "2025-11",
      "amount": 100.00
    }
  ],
  "next_billing": {
    "total_next_billing": 49.90,
    "next_billing_dates": ["2025-12-27T10:00:00Z"]
  }
}
```

---

## Library

### Get My Library
**GET** `/library`

Get all purchased products.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
[
  {
    "purchase": {
      "id": "uuid-string",
      "user_id": "uuid-string",
      "product_id": "uuid-string",
      "purchase_price": 29.99,
      "purchased_at": "2025-11-27T10:00:00Z"
    },
    "product": {
      "id": "uuid-string",
      "name": "Super Plugin",
      "type": "PLUGIN",
      "image_url": "..."
    }
  }
]
```

### Get Download URL
**GET** `/library/:id/download`

Get the temporary download URL for a purchased product.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "download_url": "https://s3.bucket/signed-url..."
}
```

---

## History

### Get History
**GET** `/history`

Get a summary of purchased products and subscriptions.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "subscriptions": [
    {
      "id": "uuid-string",
      "plan_name": "Dev Pro",
      "price_per_month": 49.90,
      "created_at": "2025-11-01T10:00:00Z"
    }
  ],
  "products": [
    {
      "id": "uuid-string",
      "name": "Super Plugin",
      "price": 29.99,
      "type": "PLUGIN"
    }
  ]
}
```

### Get Invoices
**GET** `/history/invoices`

Get invoice history.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "paid_invoices": [
    {
      "subscription_id": "uuid-string",
      "plan_name": "Dev Pro",
      "amount": 49.90,
      "due_date": "2025-11-01T10:00:00Z",
      "status": "paid"
    }
  ],
  "next_invoice": {
    "subscription_id": "uuid-string",
    "plan_name": "Dev Pro",
    "amount": 49.90,
    "due_date": "2025-12-01T10:00:00Z",
    "status": "due"
  },
  "overdue_invoices": []
}
```

---

## Subscriptions & Plans

### List Plans
**GET** `/plans`

List all active subscription plans.

**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "name": "Basic Plan",
    "description": "Start your journey",
    "price": 29.90,
    "isActive": true,
    "features": "[\"Feature 1\", \"Feature 2\"]"
  }
]
```

### List User Subscriptions
**GET** `/subscriptions`

List all subscriptions for the authenticated user.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "planName": "Dev Pro",
    "pricePerMonth": 49.90,
    "status": "ACTIVE",
    "startedAt": "2025-11-01T10:00:00Z",
    "nextBillingDate": "2025-12-01T10:00:00Z",
    "projectStage": "Desenvolvimento",
    "logs": []
  }
]
```

### Get Subscription Details
**GET** `/subscriptions/:id`

Get detailed information about a specific subscription, including project logs.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
{
  "subscription": {
    "id": "uuid-string",
    "planName": "Dev Pro",
    "pricePerMonth": 49.90,
    "status": "ACTIVE",
    "projectStage": "Desenvolvimento",
    "nextBillingDate": "2025-12-01T10:00:00Z",
    "plan": {
        "id": "uuid-string",
        "name": "Dev Pro",
        "features": "..."
    },
    "user": {
        "id": "uuid-string",
        "full_name": "John Doe",
        "email": "john@example.com",
        "discord_handle": "john#1234"
    }
  },
  "logs": [
    {
      "id": "uuid-string",
      "message": "Project started",
      "createdAt": "2025-11-01T10:00:00Z"
    }
  ]
}
```

---

### Send Chat Message
**POST** `/subscriptions/:id/chat`

Send a message to the subscription support chat.

**Headers:**
`Authorization: Bearer <token>`

**Request Body:**
```json
{
  "content": "Hello, I need help with my project."
}
```

**Response (201 Created):**
```json
{
  "id": "uuid-string",
  "subscriptionId": "uuid-string",
  "userId": "uuid-string",
  "content": "Hello, I need help with my project.",
  "isAdmin": false,
  "createdAt": "2025-11-28T12:00:00Z"
}
```

### Get Chat Messages
**GET** `/subscriptions/:id/chat`

Get the chat history for a subscription.

**Headers:**
`Authorization: Bearer <token>`

**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "subscriptionId": "uuid-string",
    "userId": "uuid-string",
    "content": "Hello, I need help with my project.",
    "isAdmin": false,
    "createdAt": "2025-11-28T12:00:00Z"
  },
  {
    "id": "uuid-string",
    "subscriptionId": "uuid-string",
    "userId": "uuid-string",
    "content": "Hi! How can we help you?",
    "isAdmin": true,
    "createdAt": "2025-11-28T12:05:00Z"
  }
]
```

---

## Admin

### Get Admin Stats
**GET** `/admin/stats`

Get analytics snapshot for the admin dashboard.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Response (200 OK):**
```json
{
  "totalRevenue": 45678.90,
  "totalUsers": 1234,
  "activeProducts": 89,
  "totalSales": 456,
  "revenueGrowthPct": 23.5,
  "usersGrowthPct": 18.2,
  "productsStatus": "Estável",
  "salesGrowthPct": 15.8,
  "lastUpdated": "2025-11-27T10:00:00Z"
}
```

### Get Recent Orders
**GET** `/admin/orders/recent`

Get the 5 most recent orders.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Response (200 OK):**
```json
[
  {
    "userName": "John Doe",
    "productName": "Super Plugin",
    "value": 29.99,
    "status": "COMPLETED"
  }
]
```

### Get Top Products
**GET** `/admin/products/top`

Get the top 3 best-selling products.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Response (200 OK):**
```json
[
  {
    "productName": "Super Plugin",
    "salesCount": 150,
    "totalRevenue": 4498.50
  }
]
```

### Get Active Subscriptions
**GET** `/admin/subscriptions/active`

List all active subscriptions with user and plan details.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "userId": "uuid-string",
    "userName": "John Doe",
    "userEmail": "john@example.com",
    "planName": "Dev Pro",
    "price": 49.90,
    "status": "ACTIVE",
    "projectStage": "Desenvolvimento",
    "nextBillingDate": "2025-12-01T10:00:00Z"
  }
]
```

### Update Subscription (Admin)
**PUT** `/admin/subscriptions/:id`

Update subscription status, project stage, or billing date.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Request Body:**
```json
{
  "status": "ACTIVE",
  "projectStage": "Finalização",
  "nextBillingDate": "2025-12-15T00:00:00Z"
}
```

**Response (200 OK):**
```json
{
  "message": "Subscription updated successfully"
}
```

### Add Project Log (Admin)
**POST** `/admin/subscriptions/:id/logs`

Add a new log entry to a subscription's project history.

**Headers:**
`Authorization: Bearer <token>` (Must be an admin user)

**Request Body:**
```json
{
  "message": "Deployment to staging environment completed."
}
```

**Response (201 Created):**
```json
{
  "message": "Log added successfully"
}
```
