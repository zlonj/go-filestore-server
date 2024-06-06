CREATE TABLE `table_file` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `file_sha1` char(40) NOT NULL DEFAULT '' COMMENT 'file hash',
  `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT 'file name',
  `file_size` bigint(20) DEFAULT '0' COMMENT 'file size',
  `file_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT 'file address/location',
  `create_at` datetime default NOW() COMMENT 'creation timestamp',
  `update_at` datetime default NOW() on update current_timestamp() COMMENT 'update timestamp',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT 'status(available/disabled/deleted)',
  `ext1` int(11) DEFAULT '0' COMMENT 'backup 1',
  `ext2` text COMMENT 'backup 2',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_file_hash` (`file_sha1`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `table_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL DEFAULT '' COMMENT 'user name',
  `user_pwd` varchar(256) NOT NULL DEFAULT '' COMMENT 'encoded password',
  `email` varchar(64) DEFAULT '' COMMENT 'email',
  `phone` varchar(128) DEFAULT '' COMMENT 'phone',
  `email_validated` tinyint(1) DEFAULT 0 COMMENT 'verified email',
  `phone_validated` tinyint(1) DEFAULT 0 COMMENT 'verified phone',
  `signup_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT 'sign up timestamp',
  `last_active` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'last active timestamp',
  `profile` text COMMENT 'profile',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT 'account status(active/disabled/locked/deleted)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`user_name`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4;

CREATE TABLE `table_user_token` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(64) NOT NULL DEFAULT '' COMMENT 'account',
  `user_token` char(40) NOT NULL DEFAULT '' COMMENT 'account login token',
    PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
