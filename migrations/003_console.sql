CREATE TABLE console_users (
    id uuid PRIMARY KEY,
    username text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    real_name text NOT NULL,
    phone text NOT NULL DEFAULT '',
    role integer NOT NULL DEFAULT 1 CHECK (role IN (0, 1)),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_drones (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    model_class integer NOT NULL CHECK (model_class IN (1, 2)),
    commissioned_on date,
    serial_number text NOT NULL DEFAULT '',
    endpoint text NOT NULL DEFAULT '',
    home_zone text NOT NULL DEFAULT '',
    owner_name text NOT NULL DEFAULT '',
    owner_phone text NOT NULL DEFAULT '',
    telemetry_status integer NOT NULL DEFAULT 1 CHECK (telemetry_status BETWEEN 1 AND 3),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_operators (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    gender integer NOT NULL CHECK (gender IN (1, 2)),
    phone text NOT NULL DEFAULT '',
    skills text NOT NULL DEFAULT '',
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_capabilities (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    price numeric(10,2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    duration integer NOT NULL DEFAULT 60 CHECK (duration BETWEEN 1 AND 1440),
    status integer NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE console_commands (
    id uuid PRIMARY KEY,
    command_no text NOT NULL UNIQUE,
    drone_id uuid NOT NULL REFERENCES console_drones(id) ON DELETE RESTRICT,
    capability_id uuid NOT NULL REFERENCES console_capabilities(id) ON DELETE RESTRICT,
    operator_id uuid REFERENCES console_operators(id) ON DELETE SET NULL,
    appointment_time timestamptz,
    status integer NOT NULL DEFAULT 0 CHECK (status BETWEEN 0 AND 3),
    remark text NOT NULL DEFAULT '',
    version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_telemetry_records (
    id uuid PRIMARY KEY,
    drone_id uuid NOT NULL REFERENCES console_drones(id) ON DELETE RESTRICT,
    battery_level numeric(5,1),
    motor_temperature numeric(5,1),
    network_latency_ms numeric(8,1),
    localization_error numeric(8,3),
    joint_load numeric(8,3),
    remark text NOT NULL DEFAULT '',
    recorded_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE console_logs (
    id uuid PRIMARY KEY,
    username text NOT NULL,
    operation text NOT NULL,
    method text NOT NULL,
    ip text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX console_drones_search_idx ON console_drones(name, serial_number) WHERE deleted_at IS NULL;
CREATE INDEX console_operators_search_idx ON console_operators(name, phone) WHERE deleted_at IS NULL;
CREATE INDEX console_capabilities_search_idx ON console_capabilities(name) WHERE deleted_at IS NULL;
CREATE INDEX console_commands_status_time_idx ON console_commands(status, appointment_time DESC);
CREATE INDEX console_telemetry_drone_time_idx ON console_telemetry_records(drone_id, recorded_at DESC);
CREATE INDEX console_logs_time_idx ON console_logs(created_at DESC);

INSERT INTO console_users(id, username, password_hash, real_name, phone)
VALUES ('10000000-0000-0000-0000-000000000001', 'admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', '系统管理员', '13800138000');

INSERT INTO console_drones(id, name, model_class, commissioned_on, serial_number, endpoint, home_zone, owner_name, owner_phone, telemetry_status) VALUES
('20000000-0000-0000-0000-000000000001', '天巡一号', 1, '2024-03-15', 'UAV-BJ-0001', 'mqtt://10.20.1.11', '北京东部起降场', '北区机队', '010-60010001', 1),
('20000000-0000-0000-0000-000000000002', '云翼二号', 2, '2024-07-22', 'UAV-BJ-0002', 'mqtt://10.20.1.12', '北京西部起降场', '西区机队', '010-60010002', 2),
('20000000-0000-0000-0000-000000000003', '海东青', 1, '2025-01-08', 'UAV-BJ-0003', 'mqtt://10.20.1.13', '北京北部起降场', '北区机队', '010-60010003', 1),
('20000000-0000-0000-0000-000000000004', '星图四号', 2, '2025-05-20', 'UAV-BJ-0004', 'mqtt://10.20.1.14', '北京南部起降场', '南区机队', '010-60010004', 3),
('20000000-0000-0000-0000-000000000005', '长风五号', 1, '2025-09-12', 'UAV-BJ-0005', 'mqtt://10.20.1.15', '北京东部起降场', '东区机队', '010-60010005', 2);

INSERT INTO console_operators(id, name, gender, phone, skills) VALUES
('30000000-0000-0000-0000-000000000001', '王美玲', 2, '13800138001', '航线调度、气象研判、遥测监控'),
('30000000-0000-0000-0000-000000000002', '刘建军', 1, '13800138002', '设备维护、电池安全、故障隔离'),
('30000000-0000-0000-0000-000000000003', '陈晓红', 2, '13800138003', '任务规划、空域协调、地面站操作'),
('30000000-0000-0000-0000-000000000004', '张伟强', 1, '13800138004', '应急返航、安全复核、事故响应');

INSERT INTO console_capabilities(id, name, description, price, duration) VALUES
('40000000-0000-0000-0000-000000000001', '常规设施巡检', '按预定航线采集设施影像与遥测数据', 100.00, 60),
('40000000-0000-0000-0000-000000000002', '精细航线巡检', '由飞行调度员执行的高精度航线与设备巡检', 150.00, 45),
('40000000-0000-0000-0000-000000000003', '风险指标观察', '记录飞行任务过程中的基础风险观察结果', 50.00, 30),
('40000000-0000-0000-0000-000000000004', '应急空域侦察', '在安全策略约束下执行应急侦察与返航', 200.00, 60);

INSERT INTO console_commands(id, command_no, drone_id, capability_id, operator_id, appointment_time, status, remark) VALUES
('50000000-0000-0000-0000-000000000001', 'UAV20260818001', '20000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', now() + interval '1 hour', 0, '等待执行'),
('50000000-0000-0000-0000-000000000002', 'UAV20260818002', '20000000-0000-0000-0000-000000000002', '40000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', now() + interval '2 hours', 1, '飞行任务进行中');

INSERT INTO console_telemetry_records(id, drone_id, battery_level, motor_temperature, network_latency_ms, localization_error, joint_load, remark, recorded_at) VALUES
('60000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 125, 82, 72, 5.6, 36.5, '状态平稳', now() - interval '2 hours'),
('60000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 135, 88, 78, 6.2, 36.4, '继续观察', now() - interval '1 hour');

INSERT INTO console_logs(id, username, operation, method, ip)
VALUES ('70000000-0000-0000-0000-000000000001', 'admin', '系统初始化', 'migration.003', '127.0.0.1');
