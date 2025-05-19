# TracePost-larvaeChain - Hướng Dẫn Sử Dụng & Quy Trình Nghiệp Vụ

## Giới thiệu

**TracePost-larvaeChain** là nền tảng truy xuất nguồn gốc tôm giống dựa trên blockchain Layer 1 tùy chỉnh, tích hợp hợp đồng thông minh, QR/NFC, lưu trữ phi tập trung (IPFS), NFT, và danh tính số (DDI). Hệ thống hướng tới minh bạch hóa chuỗi cung ứng, đáp ứng tiêu chuẩn quốc tế (EU DR, US FDA, ASC), nâng cao uy tín tôm Việt Nam.

---

## 1. Thành phần hệ thống & Công nghệ

- **Trại giống tôm**: Đăng ký, quản lý lô tôm giống.
- **Hộ nuôi tôm**: Ghi nhận dữ liệu môi trường, sự kiện nuôi.
- **Nhà máy chế biến**: Ghi nhận xử lý, đóng gói, kiểm dịch.
- **Doanh nghiệp xuất khẩu/đại lý**: Theo dõi vận chuyển, phân phối.
- **Người tiêu dùng**: Truy xuất nguồn gốc qua QR/NFT.
- **Cơ quan quản lý**: Giám sát, kiểm tra tuân thủ.

**Công nghệ nổi bật**:

- Blockchain Layer 1 (PoA/BFT), cầu nối Polkadot/Cosmos, DDI, NFT, IPFS, QR/NFC, BaaS, hợp đồng thông minh, API RESTful, JWT, Redis, PostgreSQL, Hyperledger Fabric (tùy chọn).

---

## 2. Quy trình nghiệp vụ & API chi tiết

### 2.1. Đăng ký & Xác thực

**Flow:**

1. Đăng ký tài khoản (trại giống, hộ nuôi, nhà máy, quản lý)
   - `POST /api/v1/auth/register`
2. Đăng nhập lấy JWT
   - `POST /api/v1/auth/login`
3. Làm mới token (nếu cần)
   - `POST /api/v1/auth/refresh`
4. Đăng xuất
   - `POST /api/v1/auth/logout`

---

### 2.2. API Quản trị (Admin API)

**Quản lý Người dùng:**

- Khóa/mở khóa tài khoản người dùng
  - `PUT /api/v1/admin/users/{userId}/status`
- Xem danh sách người dùng theo vai trò
  - `GET /api/v1/admin/users?role=hatchery_manager`

**Quản lý Trại giống:**

- Phê duyệt tài khoản trại giống
  - `PUT /api/v1/admin/hatcheries/{hatcheryId}/approve`

**Quản lý Tuân thủ:**

- Thu hồi chứng chỉ vi phạm
  - `PUT /api/v1/admin/certificates/{docId}/revoke`
- Kiểm tra tuân thủ tiêu chuẩn FDA/ASC
  - `POST /api/v1/admin/compliance/check`
- Xuất báo cáo đa chuẩn (GS1 EPCIS, PDF)
  - `POST /api/v1/admin/compliance/export`

**Quản lý Danh tính Phi tập trung (DID):**

- Cấp/phát hành DID cho các bên
  - `POST /api/v1/admin/identity/issue`
- Thu hồi DID bị xâm phạm
  - `POST /api/v1/admin/identity/revoke`

**Tích hợp Blockchain:**

- Cấu hình node blockchain
  - `POST /api/v1/admin/blockchain/nodes/configure`
- Giám sát giao dịch đa chuỗi
  - `GET /api/v1/admin/blockchain/monitor`

Chi tiết hơn về API Quản trị có thể xem tại [tài liệu Admin API](docs/admin_api.md).

---

### 2.3. Quản lý công ty & người dùng

**Flow:**

1. Tạo công ty (admin)
   - `POST /api/v1/companies`
2. Tạo user (admin)
   - `POST /api/v1/users`
3. Lấy danh sách công ty/người dùng
   - `GET /api/v1/companies`
   - `GET /api/v1/users`
4. Cập nhật/xóa thông tin
   - `PUT /api/v1/companies/:companyId`
   - `PUT /api/v1/users/:userId`
   - `DELETE /api/v1/companies/:companyId`
   - `DELETE /api/v1/users/:userId`
5. Lấy thông tin cá nhân, đổi mật khẩu
   - `GET /api/v1/users/me`
   - `PUT /api/v1/users/me/password`

---

### 2.3. Quản lý trại giống & lô tôm

**Flow:**

1. Tạo trại giống (admin/hatchery_manager)
   - `POST /api/v1/hatcheries`
2. Tạo lô tôm mới
   - `POST /api/v1/batches`
3. Lấy danh sách, chi tiết trại/lô
   - `GET /api/v1/hatcheries`
   - `GET /api/v1/batches`
   - `GET /api/v1/hatcheries/:hatcheryId`
   - `GET /api/v1/batches/:batchId`
4. Cập nhật/xóa trại/lô
   - `PUT /api/v1/hatcheries/:hatcheryId`
   - `PUT /api/v1/batches/:batchId`
   - `DELETE /api/v1/hatcheries/:hatcheryId`
   - `DELETE /api/v1/batches/:batchId`

---

### 2.4. Ghi nhận sự kiện, môi trường, tài liệu

**Flow:**

1. Ghi nhận sự kiện (feeding, vaccination, harvest, etc.)
   - `POST /api/v1/events`
2. Ghi nhận dữ liệu môi trường (nhiệt độ, pH, DO, v.v.)
   - `POST /api/v1/environment`
3. Tải lên tài liệu (chứng nhận, kiểm dịch)
   - `POST /api/v1/documents`
4. Lấy tài liệu, lịch sử sự kiện
   - `GET /api/v1/documents/:documentId`
   - `GET /api/v1/batches/:batchId/events`
   - `GET /api/v1/batches/:batchId/history`

---

### 2.5. Chuyển giao lô hàng (Shipment Transfer)

**Flow:**

1. Tạo giao dịch chuyển giao (trại giống → hộ nuôi, hộ nuôi → nhà máy, ...)
   - `POST /api/v1/shipments/transfers`
2. Lấy danh sách, chi tiết giao dịch
   - `GET /api/v1/shipments/transfers`
   - `GET /api/v1/shipments/transfers/:id`
   - `GET /api/v1/shipments/transfers/batch/:batchId`
3. Cập nhật/xóa giao dịch
   - `PUT /api/v1/shipments/transfers/:id`
   - `DELETE /api/v1/shipments/transfers/:id`

---

### 2.6. NFT & QR Code cho truy xuất nguồn gốc

**Flow:**

1. Mỗi lần chuyển giao, hệ thống tạo NFT đại diện cho lô hàng
   - `POST /api/v1/nft/mint` (hoặc tự động khi POST shipment transfer)
2. Gắn mã QR cho mỗi NFT/lô hàng
   - `GET /api/v1/supplychain/:batchId/qr`
   - `GET /api/v1/qr/:batchId`
3. Người dùng/đối tác quét QR để truy xuất lịch sử, xác thực nguồn gốc
   - `GET /api/v1/qr/:batchId`
   - `GET /api/v1/qr/gateway/:batchId`

---

### 2.7. Truy vết chuỗi cung ứng & kiểm tra minh bạch

**Flow:**

1. Lấy thông tin chuỗi cung ứng của lô hàng
   - `GET /api/v1/supplychain/:batchId`
2. Truy vết từng node giao dịch, kiểm tra lịch sử NFT
   - `GET /api/v1/batches/:batchId/history`
   - `GET /api/v1/nft/:nftId/history`
3. Ngăn chặn giao dịch thứ cấp không minh bạch bằng xác thực DDI & kiểm tra quyền sở hữu NFT

---

### 2.8. Blockchain & Interoperability

**Flow:**

1. Ghi nhận sự kiện lên blockchain
   - Tự động khi POST event, shipment, document, v.v.
2. Truy vấn dữ liệu on-chain
   - `GET /api/v1/blockchain/batch/:batchId`
   - `GET /api/v1/blockchain/event/:eventId`
3. Tích hợp cầu nối (bridge) với blockchain quốc tế (Cosmos, Polkadot, GS1 EPCIS)
   - `POST /api/v1/interop/bridges/cosmos`
   - `POST /api/v1/interop/bridges/polkadot`
   - `POST /api/v1/interop/ibc/send`
   - `POST /api/v1/interop/xcm/send`
   - `GET /api/v1/interop/txs/:txId`

---

### 2.9. Danh tính số phi tập trung (DDI)

**Flow:**

1. Tạo DID cho tổ chức/cá nhân
   - `POST /api/v1/identity/did`
2. Xác thực, kiểm tra DID
   - `GET /api/v1/identity/did/:did`
   - `POST /api/v1/identity/verify`
3. Tạo, xác thực claim (JWT)
   - `POST /api/v1/identity/claim`
   - `GET /api/v1/identity/claim/:claimId`

---

### 2.10. Blockchain-as-a-Service (BaaS)

**Flow:**

1. Tạo mạng blockchain mới
   - `POST /api/v1/baas/networks`
2. Quản lý mạng, triển khai hợp đồng thông minh
   - `GET /api/v1/baas/networks`
   - `POST /api/v1/baas/deployments`
   - `GET /api/v1/baas/deployments/:deploymentId`

---

### 2.11. Mobile Endpoints

**Flow:**

1. Truy xuất nguồn gốc qua QR trên mobile
   - `GET /api/v1/mobile/trace/:qrCode`
2. Lấy thông tin tóm tắt lô hàng
   - `GET /api/v1/mobile/batch/:batchId/summary`

---

## 3. Lưu ý triển khai & mở rộng

- **Tích hợp Interoperability**: Ưu tiên Cosmos/Polkadot bridge để mở rộng xuất khẩu.
- **Thử nghiệm DDI**: Áp dụng cho nhóm nhỏ trước khi mở rộng toàn hệ thống.
- **Nâng cấp đồng thuận**: Thử nghiệm PoS/sharding trong môi trường testnet.
- **Khám phá BaaS**: Đánh giá Hyperledger Fabric/IBM Blockchain cho các module bổ sung.
- **Tham gia liên minh blockchain**: Hợp tác với các ngành liên quan để tăng tiêu chuẩn hóa.

---

## 4. Ví dụ flow nghiệp vụ thực tế

### 4.1. Đăng ký trại giống mới & tạo lô tôm

1. Admin đăng ký tài khoản → Đăng nhập lấy JWT.
2. Tạo công ty trại giống → Tạo user hatchery manager.
3. Hatchery manager đăng nhập → Tạo trại giống.
4. Tạo lô tôm mới (batch) cho trại giống.
5. Ghi nhận sự kiện, môi trường, tài liệu cho lô tôm.
6. Khi chuyển giao lô tôm cho hộ nuôi: Tạo shipment transfer → Hệ thống tự động mint NFT, gắn QR.
7. Hộ nuôi quét QR kiểm tra nguồn gốc, xác thực NFT.

### 4.2. Truy xuất nguồn gốc cho người tiêu dùng

1. Người tiêu dùng quét QR trên sản phẩm.
2. Hệ thống trả về lịch sử lô hàng, các node giao dịch, chứng nhận, NFT, xác thực blockchain.

---

## 5. Tài liệu tham khảo & tiêu chuẩn

- [EU Digital Product Passport](https://ec.europa.eu/)
- [US FDA Food Traceability](https://www.fda.gov/)
- [GS1 EPCIS Standard](https://www.gs1.org/standards/epcis)
- [W3C Decentralized Identifiers (DID)](https://www.w3.org/TR/did-core/)
- [Polkadot, Cosmos Interoperability](https://polkadot.network/), (https://cosmos.network/)

---

## 6. Liên hệ & hỗ trợ

- Email: support@vietchain.com
- Tài liệu API: `/swagger/`
- Đội ngũ phát triển: TracePost-larvaeChain Team

---

**Lưu ý:**  
Mọi API đều yêu cầu JWT (trừ các endpoint public như truy xuất QR). Để tích hợp, hãy thực hiện đúng trình tự flow nghiệp vụ như trên để đảm bảo dữ liệu xuyên suốt, minh bạch và truy xuất được toàn bộ chuỗi cung ứng.

---

## 7. Hướng dẫn sử dụng API TracePost-larvaeChain trên Postman

### 7.1. Import collection (nếu có)

- Nếu dự án cung cấp file Postman collection (`.json`), vào Postman > Import > Chọn file collection để import toàn bộ API mẫu.
- Nếu không có, bạn có thể tự tạo request mới theo từng endpoint hướng dẫn ở trên.

### 7.2. Thiết lập môi trường (Environment)

- Tạo Environment mới, đặt biến `base_url` (ví dụ: `http://localhost:8080` hoặc domain thực tế).
- Thêm biến `jwt_token` để lưu JWT sau khi đăng nhập.

### 7.3. Lấy JWT (Đăng nhập)

- Tạo request `POST {{base_url}}/api/v1/auth/login` với body dạng JSON:
  ```json
  {
    "email": "your_email@example.com",
    "password": "your_password"
  }
  ```
- Sau khi login thành công, copy giá trị `access_token` trả về, gán vào biến `jwt_token` của environment.

### 7.4. Gán JWT vào header Authorization

- Ở mỗi request cần xác thực, vào tab Authorization > chọn Bearer Token > điền `{{jwt_token}}`.
- Hoặc thêm header thủ công:
  ```
  Authorization: Bearer {{jwt_token}}
  ```

### 7.5. Gửi request mẫu

- Đăng ký tài khoản: `POST {{base_url}}/api/v1/auth/register`
- Tạo công ty: `POST {{base_url}}/api/v1/companies`
- Tạo lô tôm: `POST {{base_url}}/api/v1/batches`
- Truy xuất QR: `GET {{base_url}}/api/v1/qr/:batchId`
- Mint NFT: `POST {{base_url}}/api/v1/nft/mint`

### 7.6. Lưu ý khi sử dụng Postman

- Đảm bảo luôn gửi đúng header `Content-Type: application/json` cho các request POST/PUT.
- Nếu gặp lỗi 401/403, kiểm tra lại JWT hoặc quyền user.
- Các endpoint public (truy xuất QR) không cần JWT.
- Có thể dùng tab Tests trong Postman để tự động lưu JWT vào biến môi trường sau khi login:
  ```javascript
  // Tab Tests của request login
  var data = pm.response.json();
  pm.environment.set("jwt_token", data.access_token);
  ```

---

## 8. Hướng dẫn từng bước sử dụng API cho người dùng mới

### Bước 1: Đăng ký tài khoản

- Endpoint: `POST /api/v1/auth/register`
- Body mẫu:
  ```json
  {
    "email": "user@example.com",
    "password": "your_password",
    "role": "admin"
  }
  ```
- Lưu ý: Nếu là admin hệ thống, đăng ký với role `admin`. Các vai trò khác như `hatchery_manager`, `farmer`, ...

### Bước 2: Đăng nhập lấy JWT

- Endpoint: `POST /api/v1/auth/login`
- Body mẫu:
  ```json
  {
    "email": "user@example.com",
    "password": "your_password"
  }
  ```
- Kết quả trả về: `access_token` (JWT). Lưu lại token này để dùng cho các request tiếp theo.

### Bước 3: Tạo công ty (nếu là admin)

- Endpoint: `POST /api/v1/companies`
- Header: `Authorization: Bearer <access_token>`
- Body mẫu:
  ```json
  {
    "name": "Công ty TNHH Trại Giống ABC",
    "type": "hatchery",
    "location": "Bạc Liêu",
    "contact_info": "0123456789"
  }
  ```

### Bước 4: Tạo user cho trại giống/hatchery manager

- Endpoint: `POST /api/v1/users`
- Header: `Authorization: Bearer <access_token>`
- Body mẫu:
  ```json
  {
    "email": "hatchery1@example.com",
    "password": "your_password",
    "role": "hatchery_manager",
    "company_id": 1
  }
  ```

### Bước 5: Tạo trại giống

- Endpoint: `POST /api/v1/hatcheries`
- Header: `Authorization: Bearer <access_token>`
- Body mẫu:
  ```json
  {
    "name": "Trại giống A",
    "location": "Bạc Liêu",
    "contact": "0123456789",
    "company_id": 1
  }
  ```

### Bước 6: Tạo lô tôm mới

- Endpoint: `POST /api/v1/batches`
- Header: `Authorization: Bearer <access_token>`
- Body mẫu:
  ```json
  {
    "hatchery_id": 1,
    "species": "Penaeus monodon",
    "quantity": 100000,
    "status": "active"
  }
  ```

### Bước 7: Ghi nhận sự kiện, môi trường, tài liệu

- Ghi nhận sự kiện:
  - `POST /api/v1/events`
  - Body:
    ```json
    {
      "batch_id": 1,
      "event_type": "feeding",
      "actor_id": 2,
      "location": "Bạc Liêu",
      "metadata": { "note": "Cho ăn lần 1" }
    }
    ```
- Ghi nhận môi trường:
  - `POST /api/v1/environment`
  - Body:
    ```json
    {
      "batch_id": 1,
      "temperature": 28.5,
      "ph": 7.8,
      "salinity": 15.2,
      "density": 25.3,
      "age": 15
    }
    ```
- Upload tài liệu:
  - `POST /api/v1/documents` (multipart/form-data)
  - Form fields:
    ```json
    {
      "batch_id": "1",
      "doc_type": "certificate",
      "uploaded_by": "2",
      "file": "(file binary data)"
    }
    ```

### Bước 8: Chuyển giao lô hàng

- Endpoint: `POST /api/v1/shipments/transfers`
- Header: `Authorization: Bearer <access_token>`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "sender_id": 2,
    "receiver_id": 3,
    "status": "pending"
  }
  ```

### Bước 9: Mint NFT, lấy QR code

- Mint NFT: `POST /api/v1/nft/mint` (hoặc tự động khi chuyển giao)
- Lấy QR: `GET /api/v1/qr/:batchId` hoặc `GET /api/v1/supplychain/:batchId/qr`

### Bước 10: Truy xuất nguồn gốc bằng QR

- Người dùng hoặc đối tác quét QR, gửi request:
  - `GET /api/v1/qr/:batchId` (public, không cần JWT)
  - Hoặc trên mobile: `GET /api/v1/mobile/trace/:qrCode`

---

**Lưu ý:**

- Luôn gửi header `Authorization: Bearer <access_token>` cho các API cần xác thực.
- Tham khảo thêm chi tiết schema, response tại `/swagger/`.
- Nếu gặp lỗi 401/403, kiểm tra lại JWT hoặc quyền user.

---

## PHỤ LỤC: Danh sách đầy đủ các API endpoint TracePost-larvaeChain

**Lưu ý:** Tất cả các endpoint đều có tiền tố `/api/v1/`.

### 1. Authentication & User

- POST `/auth/register` : Đăng ký tài khoản
- POST `/auth/login` : Đăng nhập lấy JWT
- POST `/auth/refresh` : Làm mới JWT
- POST `/auth/logout` : Đăng xuất
- GET `/users` : Danh sách user
- POST `/users` : Tạo user mới
- GET `/users/me` : Thông tin cá nhân
- PUT `/users/me/password` : Đổi mật khẩu
- GET `/users/:userId` : Chi tiết user
- PUT `/users/:userId` : Cập nhật user
- DELETE `/users/:userId` : Xóa user

### 2. Company & Hatchery

- GET `/companies` : Danh sách công ty
- POST `/companies` : Tạo công ty
- GET `/companies/:companyId` : Chi tiết công ty
- PUT `/companies/:companyId` : Cập nhật công ty
- DELETE `/companies/:companyId` : Xóa công ty
- GET `/hatcheries` : Danh sách trại giống
- POST `/hatcheries` : Tạo trại giống
- GET `/hatcheries/:hatcheryId`: Chi tiết trại giống
- PUT `/hatcheries/:hatcheryId`: Cập nhật trại giống
- DELETE `/hatcheries/:hatcheryId`: Xóa trại giống

### 3. Batch (Lô tôm)

- GET `/batches` : Danh sách lô tôm
- POST `/batches` : Tạo lô tôm
- GET `/batches/:batchId` : Chi tiết lô tôm
- PUT `/batches/:batchId` : Cập nhật lô tôm
- DELETE `/batches/:batchId` : Xóa lô tôm
- GET `/batches/:batchId/events`: Lịch sử sự kiện lô
- GET `/batches/:batchId/environment`: Dữ liệu môi trường
- GET `/batches/:batchId/documents` : Danh sách tài liệu
- GET `/batches/:batchId/history` : Lịch sử truy vết
- GET `/batches/:batchId/qr` : QR code xác thực blockchain

### 4. Event, Environment, Document

- POST `/events` : Ghi nhận sự kiện
- POST `/environment` : Ghi nhận dữ liệu môi trường
- POST `/documents` : Upload tài liệu
- GET `/documents/:documentId` : Chi tiết tài liệu

### 5. Shipment Transfer

- POST `/shipments/transfers` : Tạo giao dịch chuyển giao
- GET `/shipments/transfers` : Danh sách giao dịch
- GET `/shipments/transfers/:id` : Chi tiết giao dịch
- GET `/shipments/transfers/batch/:batchId` : Giao dịch theo lô
- PUT `/shipments/transfers/:id` : Cập nhật giao dịch
- DELETE `/shipments/transfers/:id` : Xóa giao dịch

### 6. NFT & QR

- POST `/nft/mint` : Mint NFT cho lô hàng
- GET `/nft/batches/:batchId` : Thông tin NFT của lô
- GET `/nft/tokens/:tokenId` : Thông tin NFT token
- PUT `/nft/tokens/:tokenId/transfer` : Chuyển quyền NFT
- GET `/nft/:nftId/history` : Lịch sử NFT
- GET `/qr/:batchId` : QR code truy xuất
- GET `/qr/gateway/:batchId` : QR code gateway
- GET `/supplychain/:batchId/qr` : QR code xác thực blockchain

### 7. Supply Chain Tracing

- GET `/supplychain/:batchId` : Thông tin chuỗi cung ứng

### 8. Blockchain & Interoperability

- GET `/blockchain/batch/:batchId` : Dữ liệu lô trên blockchain
- GET `/blockchain/event/:eventId` : Dữ liệu sự kiện trên blockchain
- GET `/blockchain/document/:docId` : Dữ liệu tài liệu trên blockchain
- POST `/interop/bridges/cosmos` : Kết nối Cosmos bridge
- POST `/interop/bridges/polkadot` : Kết nối Polkadot bridge
- POST `/interop/ibc/send` : Gửi IBC
- POST `/interop/xcm/send` : Gửi XCM
- GET `/interop/txs/:txId` : Truy vấn giao dịch cross-chain
- GET `/interop/protocols` : Danh sách protocol hỗ trợ
- GET `/interop/chains` : Danh sách blockchain kết nối
- POST `/interop/chains` : Đăng ký blockchain ngoài
- POST `/interop/share-batch` : Chia sẻ lô với blockchain ngoài

### 9. Compliance

- GET `/compliance/check/:batchId` : Kiểm tra tuân thủ lô
- GET `/compliance/report/:batchId` : Báo cáo tuân thủ chi tiết
- GET `/compliance/standards` : Danh sách tiêu chuẩn
- POST `/compliance/validate` : Kiểm tra lô với tiêu chuẩn cụ thể

### 10. Identity (DID, Claim)

- POST `/identity/did` : Tạo DID
- GET `/identity/did/:did` : Tra cứu DID
- POST `/identity/claim` : Tạo claim
- GET `/identity/claim/:claimId` : Tra cứu claim
- POST `/identity/claim/verify` : Xác thực claim
- PUT `/identity/claim/:claimId/revoke` : Thu hồi claim
- GET `/identity/list` : Danh sách DID

### 11. BaaS (Blockchain as a Service)

- GET `/baas/networks` : Danh sách mạng blockchain
- POST `/baas/networks` : Tạo mạng blockchain
- GET `/baas/deployments` : Danh sách smart contract
- POST `/baas/deployments` : Deploy smart contract
- GET `/baas/deployments/:deploymentId` : Chi tiết deployment
- GET `/baas/templates` : Danh sách template blockchain

### 12. Analytics

- GET `/analytics/anomalies/:batchId` : Phát hiện bất thường lô
- GET `/analytics/timeline/:batchId` : Timeline giao dịch lô

### 13. Mobile Endpoints

- GET `/mobile/trace/:qrCode` : Truy xuất QR trên mobile
- GET `/mobile/batch/:batchId/summary` : Thông tin tóm tắt lô cho mobile

---

**Tham khảo chi tiết request/response, schema, ví dụ mẫu tại:**

- `/swagger/` (Swagger UI)
- File `docs/swagger.yaml` hoặc `docs/swagger.json`

---

### PHỤ LỤC: Body JSON mẫu cho các API chính

#### 1. Đăng ký tài khoản

- Endpoint: `POST /api/v1/auth/register`
- Body mẫu:
  ```json
  {
    "company_id": "1",
    "email": "string",
    "password": "string",
    "role": "user",
    "username": "string"
  }
  ```
  > Lưu ý: `role` có thể là `admin`, `hatchery_manager`, `farmer`, `processor`, v.v.

#### 2. Đăng nhập lấy JWT

- Endpoint: `POST /api/v1/auth/login`
- Body mẫu:
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```

#### 3. Tạo công ty

- Endpoint: `POST /api/v1/companies`
- Body mẫu:
  ```json
  {
    "name": "Công ty TNHH Tôm Việt",
    "type": "hatchery",
    "location": "Bạc Liêu",
    "contact_info": "0123456789",
    "is_active": true
  }
  ```

#### 4. Tạo user cho công ty

- Endpoint: `POST /api/v1/users`
- Body mẫu:
  ```json
  {
    "email": "hatchery1@example.com",
    "password": "your_password",
    "role": "hatchery_manager",
    "company_id": 1
  }
  ```

#### 5. Tạo trại giống

- Endpoint: `POST /api/v1/hatcheries`
- Body mẫu:
  ```json
  {
    "name": "Trại giống A",
    "location": "Bạc Liêu",
    "contact": "0123456789",
    "company_id": 1
  }
  ```

#### 6. Tạo lô tôm mới

- Endpoint: `POST /api/v1/batches`
- Body mẫu:
  ```json
  {
    "hatchery_id": 1,
    "species": "Penaeus monodon",
    "quantity": 100000,
    "status": "active"
  }
  ```

#### 7. Ghi nhận sự kiện

- Endpoint: `POST /api/v1/events`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "event_type": "feeding",
    "actor_id": 2,
    "location": "Bạc Liêu",
    "metadata": { "note": "Cho ăn lần 1" }
  }
  ```

#### 8. Ghi nhận môi trường

- Endpoint: `POST /api/v1/environment`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "temperature": 28.5,
    "ph": 7.8,
    "salinity": 15.2,
    "dissolved_oxygen": 6.5
  }
  ```

#### 9. Upload tài liệu

- Endpoint: `POST /api/v1/documents`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "document_type": "certificate",
    "title": "Giấy chứng nhận kiểm dịch",
    "file_url": "https://ipfs.io/ipfs/xxx",
    "issued_date": "2025-05-10",
    "issuer": "Cơ quan thú y"
  }
  ```

#### 10. Chuyển giao lô hàng

- Endpoint: `POST /api/v1/shipments/transfers`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "sender_id": 2,
    "receiver_id": 3,
    "status": "pending"
  }
  ```

#### 11. Mint NFT cho lô hàng

- Endpoint: `POST /api/v1/nft/mint`
- Body mẫu:
  ```json
  {
    "batch_id": 1,
    "owner_id": 2,
    "metadata": {
      "description": "NFT đại diện cho lô tôm giống",
      "image": "https://ipfs.io/ipfs/xxx"
    }
  }
  ```

---
