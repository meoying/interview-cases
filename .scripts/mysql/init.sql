create database if not exists `interview_cases`;
use `interview_cases` ;
CREATE TABLE IF NOT EXISTS article_static_tab0 (id BIGINT PRIMARY KEY,article_id INTEGER NOT NULL,like_cnt INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS article_static_tab1 (id BIGINT PRIMARY KEY,article_id INTEGER NOT NULL,like_cnt INTEGER NOT NULL DEFAULT 0);