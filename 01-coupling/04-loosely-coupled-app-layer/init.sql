CREATE DATABASE loosely_coupled_app_layer;

CREATE TABLE `loosely_coupled_app_layer`.`users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `first_name` longtext NOT NULL,
  `last_name` longtext NOT NULL,
  `password_hash` longtext,
  `last_ip` longtext,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
);

CREATE TABLE `loosely_coupled_app_layer`.`emails` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `address` varchar(256) NOT NULL,
  `primary` tinyint(1) NOT NULL,
  `user_id` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_emails_address` (`address`),
  KEY `fk_users_emails` (`user_id`),
  CONSTRAINT `fk_users_emails` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
);  
