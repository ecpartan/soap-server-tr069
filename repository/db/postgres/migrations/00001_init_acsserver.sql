-- +goose Up
-- +goose StatementBegin

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET search_path = public, pg_catalog;
SET default_tablespace = '';
SET default_with_oids = false;

-- EXTENSIONS --

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;

--

-- TOC entry 226 (class 32764 OID 27796)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner:
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


-- TABLES --


-- USERS --

CREATE TABLE IF NOT EXISTS user_role (
    id UUID PRIMARY KEY,
    name VARCHAR(45) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS user_group (
    id UUID PRIMARY KEY,
    name  VARCHAR(45) NOT NULL,
    role_id UUID REFERENCES user_role(id)
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL,
    email VARCHAR(50),
    group_id UUID REFERENCES user_group(id)
);


-- PROFILES --


CREATE TABLE IF NOT EXISTS firmware (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    path VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    version VARCHAR(50),
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS config (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    path VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    version VARCHAR(50),
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ
);


CREATE TABLE IF NOT EXISTS profile (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    firmware_id UUID REFERENCES firmware(id),
    config_id UUID REFERENCES config(id)
);


-- DEVICES --


CREATE TABLE IF NOT EXISTS device (
    id UUID PRIMARY KEY,
    sn VARCHAR(45) NOT NULL,
    manufacturer VARCHAR(45) NOT NULL,
    model VARCHAR(50) NOT NULL,
    oui VARCHAR(50) NOT NULL,
    sw_version VARCHAR(50) NOT NULL,
    hw_version VARCHAR(50) NOT NULL,
    uptime BIGINT,
    status VARCHAR(50),
    datamodel VARCHAR(50),
    username VARCHAR(50),
    password VARCHAR(50),
    cr_username VARCHAR(50),
    cr_password VARCHAR(50),
    cr_url  VARCHAR(100),
    mac VARCHAR(50),
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ,
    profile_id UUID REFERENCES profile(id)
);


-- TASKS --


CREATE TABLE IF NOT EXISTS task_op (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    body VARCHAR(2000)
);

CREATE TABLE IF NOT EXISTS task (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    event_code INTEGER NOT NULL,
    once BOOLEAN NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ,
    result  VARCHAR(50) NOT NULL,
    task_op_id UUID REFERENCES task_op(id),
    device_id UUID REFERENCES device(id)
);


-- DATA --

INSERT INTO user_role (id, name, description)
VALUES (gen_random_uuid(), 'admin', 'admin');

INSERT INTO user_group (id, name, role_id)
SELECT gen_random_uuid(), 'admin', id FROM user_role WHERE name = 'admin';

INSERT INTO users (id, username, password, email, group_id)
SELECT gen_random_uuid(), 'admin', 'admin', 'admin@admin.com', id FROM user_group WHERE name = 'admin';

INSERT INTO firmware (id, name, path, size, version, created_at, updated_at)
SELECT gen_random_uuid(), 'firmware', 'firmware', 100, '1.0.0', NOW(), NOW();

INSERT INTO config (id, name, path, size, version, created_at, updated_at)
SELECT gen_random_uuid(), 'config', 'config', 100, '1.0.0', NOW(), NOW();

INSERT INTO profile (id, name, description, firmware_id, config_id)
VALUES (gen_random_uuid(), 'default', 'Default profile', (SELECT id FROM firmware WHERE name = 'firmware' LIMIT 1),( SELECT id FROM config WHERE name = 'config'));


INSERT INTO device (id, sn, manufacturer, model, oui, sw_version, hw_version, cr_url, uptime, status, datamodel, username, password, cr_username, cr_password, mac, created_at, updated_at, profile_id)
SELECT gen_random_uuid(), '94DE80BF38B2', 'D-LINK', 'DIR-825', '94DE80', 'develop', 'DebugOnHost', 'http://127.0.0.1:8999/', 0, 'off', '98', '', '', '', '', '94:DE:80:BF:38:B2', NOW(), NOW(), id FROM profile WHERE name = 'default';


INSERT INTO device (id, sn, manufacturer, model, oui, sw_version, hw_version, cr_url, uptime, status, datamodel, username, password, cr_username, cr_password, mac, created_at, updated_at, profile_id)
SELECT gen_random_uuid(), '009012BD0001', 'D-LINK', 'DWR-956', '009012', 'develop', 'B1', 'http://192.168.0.1:8999/', 0, 'off', '98', 'user1', 'pass1', 'crreqw1', 'crreqwpass1', '94:DE:80:BF:38:B2', NOW(), NOW(), id FROM profile WHERE name = 'default';


INSERT INTO task_op (id, name, body)
VALUES (gen_random_uuid(), 'Script1', '{"Script":{"1":{"GetParameterValues":{"Name":["InternetGatewayDevice.WANDevice."]}},"Serial":"94DE80BF38B2"}}');
INSERT INTO task_op (id, name, body)
VALUES (gen_random_uuid(), 'Script2', '{"Script":{"1":{"GetParameterAttributes":{"Name":["InternetGatewayDevice.WANDevice."]}},"Serial":"94DE80BF38B2"}}');
INSERT INTO task_op (id, name, body)
VALUES (gen_random_uuid(), 'Script3', '{"Script":{"1":{"GetParameterNames":{"Name":"InternetGatewayDevice.WANDevice.","NextLevel":0}},"Serial":"94DE80BF38B2"}}');

INSERT INTO task (id, type, status, event_code, once, created_at, updated_at, result, task_op_id, device_id)
VALUES (gen_random_uuid(), 'Script', 'pending', 1, false, NOW(), NOW(), 'no', (SELECT id FROM task_op WHERE name = 'Script1' LIMIT 1), (SELECT id FROM device WHERE sn = '94DE80BF38B2'));

INSERT INTO task (id, type, status, event_code, once, created_at, updated_at, result, task_op_id, device_id)
VALUES (gen_random_uuid(), 'Script', 'pending', 1, false, NOW(), NOW(), 'no', (SELECT id FROM task_op WHERE name = 'Script2' LIMIT 1), (SELECT id FROM device WHERE sn = '94DE80BF38B2'));

INSERT INTO task (id, type, status, event_code, once, created_at, updated_at, result, task_op_id, device_id)
VALUES (gen_random_uuid(), 'Script', 'pending', 1, false, NOW(), NOW(), 'no', (SELECT id FROM task_op WHERE name = 'Script3' LIMIT 1), (SELECT id FROM device WHERE sn = '94DE80BF38B2'));



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP table IF EXISTS users, user_role, user_group, task_op, task, profile, firmware, device, config;

-- +goose StatementEnd
