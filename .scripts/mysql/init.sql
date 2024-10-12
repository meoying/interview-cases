create database if not exists `interview_cases`;


-- Create delay_msg_db_0 if it does not exist
CREATE DATABASE IF NOT EXISTS delay_msg_db_0;

-- Use delay_msg_db_0
USE delay_msg_db_0;

-- Create delay_msg_tab_0 if it does not exist
CREATE TABLE IF NOT EXISTS delay_msg_tab_0
(
    id       BIGINT AUTO_INCREMENT PRIMARY KEY,
    topic    VARCHAR(512),
    value    BLOB,
    `key`  VARCHAR(512),
    deadline BIGINT,
    status   TINYINT(3) COMMENT '0-待完成 1-完成',
    ctime    BIGINT,
    utime    BIGINT,

    INDEX (deadline),
    INDEX (utime),
    UNIQUE(`key`)
);

-- Create delay_msg_tab_1 if it does not exist
CREATE TABLE IF NOT EXISTS delay_msg_tab_1
(
    id       BIGINT AUTO_INCREMENT PRIMARY KEY,
    topic    VARCHAR(512),
    value    BLOB,
    `key`  VARCHAR(512),
    deadline BIGINT,
    status   TINYINT(3) COMMENT '0-待完成 1-完成',
    ctime    BIGINT,
    utime    BIGINT,

    INDEX (deadline),
    INDEX (utime),
    UNIQUE(`key`)
);

-- Create delay_msg_db_1 if it does not exist
CREATE DATABASE IF NOT EXISTS delay_msg_db_1;

-- Use delay_msg_db_1
USE delay_msg_db_1;

-- Create delay_msg_tab_0 if it does not exist
CREATE TABLE IF NOT EXISTS delay_msg_tab_0
(
    id       BIGINT AUTO_INCREMENT PRIMARY KEY,
    topic    VARCHAR(512),
    value    BLOB,
    `key`  VARCHAR(512),
    deadline BIGINT,
    status   TINYINT(3) COMMENT '0-待完成 1-完成',
    ctime    BIGINT,
    utime    BIGINT,

    INDEX (deadline),
    INDEX (utime),
    UNIQUE(`key`)
    );

-- Create delay_msg_tab_1 if it does not exist
CREATE TABLE IF NOT EXISTS delay_msg_tab_1
(
    id       BIGINT AUTO_INCREMENT PRIMARY KEY,
    topic    VARCHAR(512),
    value    BLOB,
    `key`  VARCHAR(512),
    deadline BIGINT,
    status   TINYINT(3) COMMENT '0-待完成 1-完成',
    ctime    BIGINT,
    utime    BIGINT,

    INDEX (deadline),
    INDEX (utime),
    UNIQUE(`key`)
);


use `interview_cases` ;
CREATE TABLE IF NOT EXISTS article_static_tab0 (id BIGINT PRIMARY KEY,article_id INTEGER NOT NULL,like_cnt INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS article_static_tab1 (id BIGINT PRIMARY KEY,article_id INTEGER NOT NULL,like_cnt INTEGER NOT NULL DEFAULT 0);