# 存储配置

Athens 支持多种存储后端用于模块持久化。您可以使用 `ATHENS_STORAGE_TYPE` 环境变量配置存储后端。

## 支持的存储后端

- 磁盘: 将模块存储在本地磁盘
- 内存: 内存存储（非持久化）
- MongoDB: 将模块存储在 MongoDB 中
- AWS S3: 将模块存储在 S3 存储桶中
- Google 云存储: 将模块存储在 GCS 存储桶中

## 配置示例

### 磁盘存储
```bash
ATHENS_STORAGE_TYPE=disk
ATHENS_DISK_STORAGE_ROOT=/path/to/storage
```

### MongoDB 存储
```bash
ATHENS_STORAGE_TYPE=mongo
ATHENS_MONGO_STORAGE_URL=mongodb://localhost:27017
ATHENS_MONGO_STORAGE_DATABASE=athens
ATHENS_MONGO_STORAGE_COLLECTION=modules
```

更多详情，请参阅[存储配置参考](#)。