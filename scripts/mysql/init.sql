create database if not exists `test`;
CREATE TABLE IF NOT EXISTS `test`.`user`
(
    id       INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50),
    email    VARCHAR(100),
    age      INT,
    KEY `idx_username` (`username`)
    );

CREATE TABLE IF NOT EXISTS `test`.`products` (
   `id` INT AUTO_INCREMENT PRIMARY KEY,
   `name` VARCHAR ( 100 ),
    `price` DECIMAL ( 10, 2 ),
    `category` VARCHAR ( 50 ),
    INDEX `idx_name` ( `name` ),
    INDEX `idx_price` ( `price` )
    );