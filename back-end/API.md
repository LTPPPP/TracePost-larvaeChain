# Blockchain Logistics Traceability System Architecture

## Tổng quan về Hệ thống

Hệ thống truy xuất nguồn gốc logistics sử dụng blockchain được thiết kế để theo dõi nguồn gốc và tính toàn vẹn của sản phẩm (chủ yếu là tôm giống) trong toàn bộ chuỗi cung ứng, từ trại giống đến người tiêu dùng cuối. Hệ thống này tích hợp các công nghệ tiên tiến như blockchain, danh tính phi tập trung (DID), tokenization (NFT), và khả năng tương tác giữa các blockchain khác nhau.

## Luồng hoạt động của Hệ thống (System Flow)

Hệ thống truy xuất nguồn gốc hoạt động theo một quy trình tuần tự và có tính tích hợp cao, được phân chia thành các giai đoạn chính như sau:

### 1. Thiết lập cơ sở hạ tầng (BaaS API - Đầu tiên)

BaaS API (Blockchain-as-a-Service) là điểm khởi đầu của toàn bộ hệ thống, cung cấp cơ sở hạ tầng blockchain làm nền tảng cho mọi hoạt động:

- **Thiết lập mạng blockchain** thông qua `POST /baas/networks` để tạo mạng mới, hỗ trợ đa nền tảng (Hyperledger Fabric, Cosmos SDK, Polkadot)
- **Cấu hình nút blockchain** với các tham số kỹ thuật phù hợp với nhu cầu truy xuất nguồn gốc
- **Triển khai smart contract** cho các hoạt động chuỗi cung ứng và truy xuất nguồn gốc
- **Thiết lập các cơ chế đồng thuận (consensus)** phù hợp với yêu cầu hệ thống
- **Tạo điều kiện cho mở rộng quy mô** (scaling) khi hệ thống phát triển

**Endpoint mẫu:**

```
POST /baas/networks
GET /baas/networks
GET /baas/networks/{networkId}
```

**Kết hợp với các API khác:**

- Admin API: Quản trị mạng blockchain
- Scaling API: Tối ưu hiệu suất hệ thống
- Blockchain API: Quản lý thực tế blockchain

### 2. Quản lý Danh tính Phi tập trung (Identity API - Thứ hai)

Sau khi thiết lập cơ sở hạ tầng blockchain, Identity API đóng vai trò quan trọng trong việc đảm bảo mọi thực thể trong chuỗi cung ứng đều được xác thực và nhận diện đúng:

- **Tạo danh tính phi tập trung (DID)** cho các bên tham gia: trại giống, nông dân, nhà chế biến, nhà xuất khẩu thông qua `POST /identity/did`
- **Giải quyết và xác minh DID** để đảm bảo tính xác thực của người tham gia qua `GET /identity/did/{did}`
- **Quản lý chứng chỉ xác thực (Verifiable Credentials)** để cung cấp thông tin chi tiết về mỗi thực thể
- **Triển khai tiêu chuẩn W3C DID** để đảm bảo tính tương thích và khả năng hoạt động toàn cầu
- **Hỗ trợ ủy quyền và kiểm soát truy cập** dựa trên danh tính được xác minh

**Endpoint mẫu:**

```
POST /identity/did
GET /identity/did/{did}
PUT /identity/did/{did}/status
POST /identity/vc/issue
GET /identity/vc/verify/{credentialId}
```

**Kết hợp với các API khác:**

- Auth API: Xác thực người dùng dựa trên DID
- ZKP API: Cung cấp bằng chứng không tiết lộ thông tin
- Company API: Gắn DID cho các công ty trong chuỗi cung ứng

### 3. Quản lý Trang trại và Sản xuất (Farms API - Thứ ba)

Sau khi thiết lập danh tính, Farms API được sử dụng để ghi lại và quản lý thông tin về hoạt động sản xuất tại nguồn:

- **Đăng ký và quản lý trang trại** với thông tin chi tiết về vị trí, loại trang trại, công suất thông qua `POST /farms`
- **Ghi lại các hoạt động nuôi trồng** như cho ăn, điều trị, giám sát các thông số môi trường
- **Quản lý lô hàng tại trang trại** bao gồm việc nhận và chuyển giao lô hàng
- **Tích hợp dữ liệu địa lý** để xác định chính xác vị trí sản xuất
- **Ghi lại dữ liệu chất lượng nước và thông số môi trường** ảnh hưởng đến sản phẩm

**Endpoint mẫu:**

```
POST /farms
GET /farms/{farmId}
POST /farms/{farmId}/records
POST /farms/{farmId}/batches/receive
POST /farms/{farmId}/batches/{batchId}/transfer
```

**Kết hợp với các API khác:**

- Batch API: Quản lý các lô hàng được tạo từ trang trại
- Hatch API: Liên kết với quá trình ương giống từ trại giống
- Geo API: Ghi lại dữ liệu địa lý của trang trại

### 4. Liên kết Giữa Các Blockchain (Interoperability API - Thứ tư)

Khi hệ thống vận hành, Interoperability API đảm bảo khả năng tương tác giữa các blockchain khác nhau, mở rộng phạm vi của hệ thống:

- **Đăng ký các blockchain bên ngoài** để tích hợp dữ liệu từ nhiều nguồn thông qua `POST /interoperability/chains/register`
- **Chia sẻ thông tin lô hàng** giữa các blockchain khác nhau thông qua các bridge và protocol chuẩn
- **Thiết lập cầu nối (bridges)** cho các nền tảng blockchain phổ biến như Polkadot và Cosmos
- **Truyền tin xuyên chuỗi** thông qua XCM (Cross-Chain Messaging) và IBC (Inter-Blockchain Communication)
- **Xác minh giao dịch xuyên chuỗi** để đảm bảo tính toàn vẹn dữ liệu

**Endpoint mẫu:**

```
POST /interoperability/chains/register
POST /interoperability/batches/share
POST /interoperability/bridges/polkadot
POST /interoperability/bridges/cosmos
POST /interoperability/xcm/message
POST /interoperability/ibc/packet
GET /interoperability/transactions/verify
```

**Kết hợp với các API khác:**

- Blockchain API: Giao tiếp với các nền tảng blockchain khác nhau
- Alliance API: Kết nối với các liên minh blockchain khác
- Compliance API: Đảm bảo tuân thủ quy định khi chia sẻ dữ liệu xuyên chuỗi

### 5. Tokenization và Truy xuất Nguồn gốc (NFT API - Thứ năm)

Cuối cùng, NFT API được sử dụng để biến các tài sản vật lý thành token kỹ thuật số, tạo điều kiện cho việc chuyển quyền sở hữu và truy xuất nguồn gốc:

- **Triển khai smart contract NFT** để tokenize các lô hàng và giao dịch thông qua `POST /nft/contracts`
- **Chuyển đổi lô hàng thành NFT** để theo dõi quyền sở hữu và lịch sử thông qua `POST /nft/batches/tokenize`
- **Tạo NFT cho giao dịch vận chuyển** để xác minh tính xác thực của hàng hóa
- **Liên kết NFT với mã QR** để kết nối thế giới vật lý và kỹ thuật số
- **Truy vết lịch sử hoàn chỉnh** của sản phẩm thông qua chuỗi NFT

**Endpoint mẫu:**

```
POST /nft/contracts
POST /nft/batches/tokenize
POST /nft/transactions/tokenize
GET /nft/tokens/{tokenId}
PUT /nft/tokens/{tokenId}/transfer
GET /nft/transactions/{transferId}/trace
GET /nft/transactions/{transferId}/qr
```

**Kết hợp với các API khác:**

- Batch API: Kết nối thông tin lô hàng với NFT
- QR API: Tạo mã QR liên kết với NFT
- Shipment API: Theo dõi vận chuyển thông qua NFT

## Kết hợp Toàn bộ Hệ thống

Các thành phần của hệ thống hoạt động trong một hệ sinh thái tích hợp chặt chẽ, với nhiều API hỗ trợ cho từng giai đoạn:

### Quá trình đăng ký và Xác thực

- **Auth API** xác thực người dùng và phân quyền truy cập
- **Admin API** quản lý cấu hình hệ thống và quyền người dùng
- **Company API** quản lý thông tin doanh nghiệp trong chuỗi cung ứng

### Quản lý Sản xuất và Quy trình

- **Batch API** theo dõi từng lô hàng từ khi tạo ra đến khi đến tay người tiêu dùng
- **Hatch API** ghi lại thông tin về nguồn gốc giống tôm
- **Processor API** quản lý quá trình chế biến sản phẩm
- **Exporter API** theo dõi quá trình xuất khẩu

### Logistics và Vận chuyển

- **Shipment API** quản lý vận chuyển giữa các giai đoạn trong chuỗi cung ứng
- **Geo API** theo dõi vị trí địa lý của hàng hóa trong quá trình vận chuyển
- **QR API** tạo và quản lý mã QR để theo dõi sản phẩm

### Tương tác và Mở rộng

- **Alliance API** quản lý hợp tác giữa các tổ chức trong chuỗi cung ứng
- **Compliance API** đảm bảo tuân thủ các quy định pháp lý
- **SupplyChain API** cung cấp giao diện chung cho quản lý chuỗi cung ứng

### Phân tích và Giám sát

- **Analytics API** phân tích dữ liệu từ chuỗi cung ứng
- **Scaling API** tối ưu hóa hiệu suất của hệ thống
- **Blockchain API** truy vấn và tương tác trực tiếp với blockchain

### Bảo mật và Quyền riêng tư

- **ZKP API** cung cấp bằng chứng không tiết lộ thông tin
- **Compliance API** đảm bảo tuân thủ các quy định về bảo vệ dữ liệu

## Luồng Dữ liệu Chi tiết trong Hệ thống

Dưới đây là mô tả chi tiết về luồng dữ liệu qua các giai đoạn khác nhau trong hệ thống:

### 1. Bắt đầu quá trình (Trại giống)

1. **Đăng ký Trại giống**:

   - Trại giống đăng ký trên hệ thống thông qua Company API
   - Mỗi trại được cấp DID thông qua Identity API
   - Thông tin trại được lưu trữ trên blockchain thông qua BaaS API

2. **Tạo lô tôm giống**:

   - Trại giống tạo lô tôm mới thông qua Hatch API
   - Thông tin di truyền và nguồn gốc được ghi lại thông qua Batch API
   - Mỗi lô hàng nhận một mã định danh duy nhất được lưu trữ trên blockchain

3. **Giám sát chất lượng**:
   - Các thông số chất lượng nước và điều kiện nuôi được ghi lại
   - Thông tin về thức ăn, thuốc và xét nghiệm sức khỏe được lưu trữ
   - Các bên độc lập có thể xác minh thông tin này thông qua blockchain

### 2. Nuôi trồng và Phát triển (Trang trại)

1. **Nhận và chăm sóc**:

   - Trang trại nhận lô tôm giống thông qua Shipment API
   - Xác nhận giao dịch được lưu trữ trên blockchain
   - Toàn bộ quá trình nuôi trồng được ghi lại thông qua Farms API

2. **Theo dõi phát triển**:

   - Các hoạt động như cho ăn, điều trị được ghi lại qua Farms API
   - Thông số môi trường và tăng trưởng được theo dõi thường xuyên
   - Dữ liệu được đồng bộ hóa với blockchain thông qua BaaS API

3. **Chuyển giao lô hàng**:
   - Khi tôm đạt kích thước thích hợp, lô hàng được chuẩn bị để chuyển đi
   - Thông tin chuyển giao được ghi lại thông qua Shipment API
   - Token NFT có thể được tạo để đại diện cho lô hàng thông qua NFT API

### 3. Chế biến và Đóng gói (Nhà máy)

1. **Tiếp nhận nguyên liệu**:

   - Nhà máy nhận lô hàng và xác nhận thông qua Shipment API
   - QR code được quét để xác minh nguồn gốc thông qua QR API
   - Thông tin được xác thực thông qua blockchain

2. **Quá trình chế biến**:

   - Từng công đoạn chế biến được ghi lại thông qua Processor API
   - Các thông số như nhiệt độ, thời gian được lưu trữ
   - Thông tin về phụ gia, bảo quản được cập nhật

3. **Đóng gói và dán nhãn**:
   - Sản phẩm được đóng gói và gắn mã QR mới
   - QR code được liên kết với toàn bộ lịch sử sản phẩm
   - NFT có thể được cập nhật để phản ánh sản phẩm đã chế biến

### 4. Xuất khẩu và Vận chuyển (Nhà xuất khẩu)

1. **Chuẩn bị xuất khẩu**:

   - Tài liệu và chứng chỉ xuất khẩu được chuẩn bị thông qua Exporter API
   - Tuân thủ quy định được xác minh thông qua Compliance API
   - Thông tin được đồng bộ với blockchain thông qua Interoperability API

2. **Vận chuyển quốc tế**:

   - Điều kiện vận chuyển được giám sát thông qua Shipment API
   - Dữ liệu GPS theo dõi vị trí hàng hóa thông qua Geo API
   - Dữ liệu được cập nhật trên blockchain với blockchain ngoài (nếu cần)

3. **Thông quan và giao hàng**:
   - Thủ tục hải quan được ghi lại và xác minh
   - Việc giao hàng cho người nhận được xác nhận
   - Thông tin được cập nhật trên NFT và blockchain

### 5. Phân phối và người tiêu dùng (Bán lẻ)

1. **Tiếp nhận tại điểm bán lẻ**:

   - Cửa hàng bán lẻ xác nhận nhận hàng thông qua Shipment API
   - QR code được quét để xác minh lịch sử sản phẩm
   - Thông tin được cập nhật trên blockchain

2. **Tiếp cận người tiêu dùng**:
   - Người tiêu dùng có thể quét QR code để xem toàn bộ lịch sử sản phẩm
   - Thông tin về nguồn gốc, chất lượng, và chứng nhận được hiển thị
   - Feedback có thể được ghi lại trên blockchain

## API Endpoint Chi tiết

Dưới đây là danh sách đầy đủ các API endpoint chính trong hệ thống, được tổ chức theo thứ tự sử dụng:

### 1. BaaS API Endpoints

```
POST /baas/networks
GET /baas/networks
GET /baas/networks/{networkId}
PUT /baas/networks/{networkId}/status
POST /baas/networks/{networkId}/organizations
POST /baas/networks/{networkId}/channels
POST /baas/networks/{networkId}/contracts
GET /baas/networks/{networkId}/monitor
```

### 2. Identity API Endpoints

```
POST /identity/did
GET /identity/did/{did}
PUT /identity/did/{did}/status
POST /identity/vc/issue
GET /identity/vc/verify/{credentialId}
POST /identity/did/{did}/authenticate
GET /identity/did/{did}/document
PUT /identity/did/{did}/controller
POST /identity/did/{did}/service
DELETE /identity/did/{did}/service/{serviceId}
```

### 3. Farms API Endpoints

```
POST /farms
GET /farms
GET /farms/{farmId}
PUT /farms/{farmId}
DELETE /farms/{farmId}
POST /farms/{farmId}/records
GET /farms/{farmId}/records
POST /farms/{farmId}/batches/receive
GET /farms/{farmId}/batches
POST /farms/{farmId}/batches/{batchId}/transfer
GET /farms/{farmId}/batches/{batchId}/history
POST /farms/{farmId}/batches/{batchId}/treatments
POST /farms/{farmId}/batches/{batchId}/feedings
POST /farms/{farmId}/batches/{batchId}/monitoring
```

### 4. Interoperability API Endpoints

```
POST /interoperability/chains/register
GET /interoperability/chains
GET /interoperability/chains/{chainId}
POST /interoperability/batches/share
POST /interoperability/bridges/polkadot
POST /interoperability/bridges/cosmos
POST /interoperability/channels/ibc
GET /interoperability/channels
POST /interoperability/xcm/message
POST /interoperability/ibc/packet
GET /interoperability/transactions/{txId}
GET /interoperability/transactions/verify
```

### 5. NFT API Endpoints

```
POST /nft/contracts
GET /nft/contracts
GET /nft/contracts/{contractAddress}
POST /nft/batches/tokenize
GET /nft/batches/{batchId}
POST /nft/transactions/tokenize
GET /nft/tokens/{tokenId}
PUT /nft/tokens/{tokenId}/transfer
GET /nft/transactions/{transferId}
GET /nft/transactions/{transferId}/trace
GET /nft/transactions/{transferId}/qr
POST /nft/tokens/{tokenId}/metadata
GET /nft/owners/{address}/tokens
```

### API Hỗ trợ và Tích hợp

#### Batch API Endpoints

```
POST /batches
GET /batches
GET /batches/{batchId}
PUT /batches/{batchId}/status
GET /batches/{batchId}/history
GET /batches/{batchId}/qr
POST /batches/{batchId}/events
GET /batches/{batchId}/certificates
POST /batches/merge
POST /batches/split
```

#### Shipment API Endpoints

```
POST /shipments/transfers
GET /shipments/transfers
GET /shipments/transfers/{id}
PUT /shipments/transfers/{id}
DELETE /shipments/transfers/{id}
GET /shipments/transfers/batch/{batchId}
GET /shipments/transfers/{id}/qr
POST /shipments/transfers/{id}/confirm
GET /shipments/transfers/{id}/conditions
POST /shipments/transfers/{id}/track
GET /shipments/status
```

#### QR API Endpoints

```
GET /qr/{code}
POST /qr/generate
GET /qr/batch/{batchId}
GET /qr/shipment/{shipmentId}
GET /qr/nft/{tokenId}
POST /qr/verify
GET /qr/history/{code}
POST /qr/link
```

#### Hatch API Endpoints

```
POST /hatcheries
GET /hatcheries
GET /hatcheries/{hatcheryId}
PUT /hatcheries/{hatcheryId}
DELETE /hatcheries/{hatcheryId}
POST /hatcheries/{hatcheryId}/certify
GET /hatcheries/{hatcheryId}/certificates
POST /hatcheries/{hatcheryId}/batches
GET /hatcheries/{hatcheryId}/batches
POST /hatcheries/{hatcheryId}/genetic-info
```

#### Processor API Endpoints

```
POST /processors
GET /processors
GET /processors/{processorId}
PUT /processors/{processorId}
DELETE /processors/{processorId}
POST /processors/{processorId}/receive
POST /processors/{processorId}/process
POST /processors/{processorId}/package
GET /processors/{processorId}/batches
POST /processors/{processorId}/quality-check
```

#### Exporter API Endpoints

```
POST /exporters
GET /exporters
GET /exporters/{exporterId}
PUT /exporters/{exporterId}
DELETE /exporters/{exporterId}
POST /exporters/{exporterId}/prepare-shipment
POST /exporters/{exporterId}/documentation
GET /exporters/{exporterId}/batches
POST /exporters/{exporterId}/customs-clearance
GET /exporters/{exporterId}/certificates
```

## Các ví dụ về Luồng dữ liệu

### Ví dụ 1: Tạo lô tôm giống mới và theo dõi đến người tiêu dùng

2. **Tạo DID cho lô hàng** (Identity API → Blockchain API)

   ```json
   POST /identity/did
   {
     "entityType": "batch",
     "entityName": "Batch-LV-20250501-12345",
     "metadata": {
       "batchId": "LV-20250501-12345",
       "hatcheryId": 12345,
       "type": "shrimp_larvae"
     }
   }
   ```

3. **Chuyển lô hàng đến trang trại** (Shipment API → Blockchain API)

   ```json
   POST /shipments/transfers
   {
     "batchId": "LV-20250501-12345",
     "senderId": 12345,
     "receiverId": 56789,
     "transferTime": "2025-05-15T09:30:00Z",
     "quantity": 50000,
     "conditions": {
       "temperature": "23C",
       "oxygenLevel": "8.2mg/L",
       "containerType": "specialized_tank"
     }
   }
   ```

4. **Nhận tôm giống tại trang trại** (Farms API → Blockchain API)

   ```json
   POST /farms/56789/batches/receive
   {
     "batchId": "LV-20250501-12345",
     "receivedQuantity": 49500,
     "survivalRate": 99,
     "receivedDate": "2025-05-15T14:30:00Z",
     "notes": "500 casualties during transport, otherwise good condition",
     "pondAssignment": "Pond-A12"
   }
   ```

5. **Ghi lại hoạt động nuôi trồng** (Farms API → Blockchain API)

   ```json
   POST /farms/56789/batches/LV-20250501-12345/feedings
   {
     "date": "2025-05-20T08:00:00Z",
     "feedType": "Organic feed pellets",
     "quantity": "25kg",
     "feederId": "farm_worker_123",
     "pondId": "Pond-A12"
   }
   ```

6. **Tokenize lô hàng thành NFT khi thu hoạch** (NFT API → Blockchain API)
   ```json
   POST /nft/batches/tokenize
   {
     "batchId": "LV-20250501-12345",
     "networkId": "net-20230515123456",
     "contractAddress": "0x1234567890AbCdEf1234567890AbCdEf12345678",
     "recipientAddress": "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
     "metadata": {
       "harvestDate": "2025-07-15T06:00:00Z",
       "averageSize": "22g",
       "totalWeight": "850kg",
       "qualityGrade": "Premium",
       "certifications": ["ASC", "BAP", "Organic"]
     }
   }
   ```

### Ví dụ 2: Xác minh sản phẩm bởi người tiêu dùng cuối

1. **Người tiêu dùng quét mã QR trên sản phẩm** (QR API → Blockchain API)

   ```
   GET /qr/SP-202507-12345
   ```

2. **Hệ thống trả về toàn bộ lịch sử sản phẩm**:
   ```json
   {
     "success": true,
     "message": "Product information retrieved successfully",
     "data": {
       "productId": "SP-202507-12345",
       "productName": "Premium Organic Vannamei Shrimp",
       "origin": {
         "hatchery": "EcoCert Hatchery, Vietnam",
         "farm": "OceanFresh Aquaculture, Vietnam",
         "processor": "SeaDelights Processing, Vietnam"
       },
       "journey": [
         {
           "stage": "Hatching",
           "location": "EcoCert Hatchery, Da Nang",
           "date": "2025-05-01",
           "details": "Hatched from certified SPF broodstock"
         },
         {
           "stage": "Farming",
           "location": "OceanFresh Aquaculture, Quang Nam",
           "date": "2025-05-15 to 2025-07-15",
           "details": "Raised in certified organic ponds with sustainable practices"
         },
         {
           "stage": "Processing",
           "location": "SeaDelights Processing, Ho Chi Minh City",
           "date": "2025-07-16",
           "details": "Processed under HACCP standards"
         },
         {
           "stage": "Export",
           "from": "Vietnam",
           "to": "United States",
           "date": "2025-07-18"
         },
         {
           "stage": "Distribution",
           "location": "Seafood Market, New York",
           "date": "2025-07-25"
         }
       ],
       "certifications": ["ASC", "BAP", "Organic", "HACCP"],
       "sustainability": {
         "carbonFootprint": "12.5 kg CO2e",
         "waterUsage": "Sustainable level",
         "feedConversionRatio": 1.2
       },
       "verificationMethod": "Blockchain-verified with NFT authentication"
     }
   }
   ```

## Tổng kết

Hệ thống truy xuất nguồn gốc logistics dựa trên blockchain này cung cấp một giải pháp toàn diện để theo dõi sản phẩm từ nguồn gốc đến người tiêu dùng. Thông qua sự kết hợp của năm API chính (BaaS, Identity, Farms, Interoperability và NFT) cùng với các API hỗ trợ, hệ thống đảm bảo:

1. **Minh bạch hoàn toàn** trong chuỗi cung ứng
2. **Không thể giả mạo dữ liệu** nhờ vào công nghệ blockchain
3. **Khả năng xác minh nhanh chóng** thông qua mã QR và NFT
4. **Tích hợp đa nền tảng** thông qua Interoperability API
5. **Quản lý danh tính tin cậy** thông qua Identity API
6. **Tokenization tài sản** thông qua NFT API

Hệ thống này không chỉ giúp nâng cao niềm tin của người tiêu dùng mà còn hỗ trợ các doanh nghiệp trong việc tuân thủ quy định, quản lý chất lượng và cải thiện hiệu quả của chuỗi cung ứng.

Với thiết kế mô-đun và khả năng mở rộng, hệ thống có thể được áp dụng cho nhiều loại sản phẩm khác nhau, từ thủy sản đến nông sản, thực phẩm và các mặt hàng giá trị cao khác.
