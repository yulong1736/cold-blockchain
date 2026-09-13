-- 初始化数据库表结构

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE,
    role VARCHAR(20) NOT NULL,
    company_name VARCHAR(200),
    company_id VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_username ON users(username);

-- 商品信息表
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    product_id VARCHAR(100) UNIQUE NOT NULL,
    batch_number VARCHAR(100) NOT NULL,
    trace_id VARCHAR(255) UNIQUE NOT NULL,
    origin VARCHAR(200) NOT NULL,
    production_time TIMESTAMP NOT NULL,
    temperature_min DECIMAL(5,2),
    temperature_max DECIMAL(5,2),
    humidity_min DECIMAL(5,2),
    humidity_max DECIMAL(5,2),
    transport_company VARCHAR(200),
    destination VARCHAR(200),
    customs_info TEXT,
    producer_id INTEGER REFERENCES users(id),
    producer_name VARCHAR(200),
    blockchain_hash VARCHAR(255),
    blockchain_tx_hash VARCHAR(255),
    block_number BIGINT,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_trace_id ON products(trace_id);
CREATE INDEX idx_products_batch_number ON products(batch_number);
CREATE INDEX idx_products_producer_id ON products(producer_id);
CREATE INDEX idx_products_blockchain_hash ON products(blockchain_hash);
CREATE INDEX idx_products_is_deleted ON products(is_deleted);

-- 商品操作历史表
CREATE TABLE IF NOT EXISTS product_histories (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    trace_id VARCHAR(255) NOT NULL,
    operation_type VARCHAR(20) NOT NULL,
    old_hash VARCHAR(255),
    new_hash VARCHAR(255),
    operator_id INTEGER REFERENCES users(id),
    operator_name VARCHAR(200),
    blockchain_hash VARCHAR(255),
    blockchain_tx_hash VARCHAR(255),
    change_details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_product_histories_trace_id ON product_histories(trace_id);
CREATE INDEX idx_product_histories_product_id ON product_histories(product_id);
CREATE INDEX idx_product_histories_blockchain_hash ON product_histories(blockchain_hash);

-- 温度监控记录表
CREATE TABLE IF NOT EXISTS temperature_records (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    trace_id VARCHAR(255) NOT NULL,
    temperature DECIMAL(5,2) NOT NULL,
    humidity DECIMAL(5,2),
    location VARCHAR(200),
    record_time TIMESTAMP NOT NULL,
    warehouse_id INTEGER REFERENCES users(id),
    warehouse_name VARCHAR(200),
    is_abnormal BOOLEAN DEFAULT FALSE,
    blockchain_hash VARCHAR(255),
    blockchain_tx_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_temperature_records_trace_id ON temperature_records(trace_id);
CREATE INDEX idx_temperature_records_product_id ON temperature_records(product_id);
CREATE INDEX idx_temperature_records_record_time ON temperature_records(record_time);

-- 运输节点表
CREATE TABLE IF NOT EXISTS transport_nodes (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id),
    trace_id VARCHAR(255) NOT NULL,
    node_name VARCHAR(200) NOT NULL,
    location VARCHAR(200),
    arrival_time TIMESTAMP,
    departure_time TIMESTAMP,
    logistics_id INTEGER REFERENCES users(id),
    logistics_name VARCHAR(200),
    status VARCHAR(20) DEFAULT 'in_transit',
    blockchain_hash VARCHAR(255),
    blockchain_tx_hash VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transport_nodes_trace_id ON transport_nodes(trace_id);
CREATE INDEX idx_transport_nodes_product_id ON transport_nodes(product_id);

-- 插入示例数据（可选）
-- INSERT INTO users (username, password, email, role, company_name) VALUES
-- ('admin', '$2a$10$...', 'admin@example.com', 'regulator', '监管机构');
