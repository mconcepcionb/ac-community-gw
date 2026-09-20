-- AzerothCore login database (acore_auth) fixture for local development.
--
-- The table definitions below mirror the official AzerothCore base schema
-- (data/sql/base/db_auth/*.sql) so the gateway's read adapter and the fake
-- worldserver behave exactly like a real server. Only the seeding is local.

CREATE DATABASE IF NOT EXISTS acore_auth;
USE acore_auth;

DROP TABLE IF EXISTS `account`;
CREATE TABLE `account` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'Identifier',
  `username` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `salt` binary(32) NOT NULL,
  `verifier` binary(32) NOT NULL,
  `session_key` binary(40) DEFAULT NULL,
  `totp_secret` varbinary(128) DEFAULT NULL,
  `email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `reg_mail` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `joindate` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_ip` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '127.0.0.1',
  `last_attempt_ip` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '127.0.0.1',
  `failed_logins` int unsigned NOT NULL DEFAULT '0',
  `locked` tinyint unsigned NOT NULL DEFAULT '0',
  `lock_country` varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '00',
  `last_login` timestamp NULL DEFAULT NULL,
  `online` int unsigned NOT NULL DEFAULT '0',
  `expansion` tinyint unsigned NOT NULL DEFAULT '2',
  `Flags` int unsigned NOT NULL DEFAULT '0',
  `mutetime` bigint NOT NULL DEFAULT '0',
  `mutereason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `muteby` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `locale` tinyint unsigned NOT NULL DEFAULT '0',
  `os` varchar(3) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `recruiter` int unsigned NOT NULL DEFAULT '0',
  `totaltime` int unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Account System';

DROP TABLE IF EXISTS `account_access`;
CREATE TABLE `account_access` (
  `id` int unsigned NOT NULL,
  `gmlevel` tinyint unsigned NOT NULL,
  `RealmID` int NOT NULL DEFAULT '-1',
  `comment` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT '',
  PRIMARY KEY (`id`,`RealmID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

DROP TABLE IF EXISTS `account_banned`;
CREATE TABLE `account_banned` (
  `id` int unsigned NOT NULL DEFAULT '0' COMMENT 'Account id',
  `bandate` int unsigned NOT NULL DEFAULT '0',
  `unbandate` int unsigned NOT NULL DEFAULT '0',
  `bannedby` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `banreason` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `active` tinyint unsigned NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`,`bandate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Ban List';

-- Seed accounts. Salt/verifier are zeroed: these accounts are never used to log
-- into a real worldserver, only to exercise the gateway's read path.
INSERT INTO `account` (username, salt, verifier, email, reg_mail, expansion, online, last_ip)
VALUES
  ('ADMIN',  UNHEX(REPEAT('00', 32)), UNHEX(REPEAT('00', 32)), 'admin@example.com',  'admin@example.com',  2, 0, '127.0.0.1'),
  ('PLAYER', UNHEX(REPEAT('00', 32)), UNHEX(REPEAT('00', 32)), 'player@example.com', 'player@example.com', 2, 0, '127.0.0.1'),
  ('BANNED', UNHEX(REPEAT('00', 32)), UNHEX(REPEAT('00', 32)), 'banned@example.com', 'banned@example.com', 2, 0, '127.0.0.1')
ON DUPLICATE KEY UPDATE email = VALUES(email);

-- GM levels: the adapter takes MAX(gmlevel), so prove both all-realm and
-- realm-specific rows work.
INSERT INTO `account_access` (id, gmlevel, RealmID, comment)
SELECT id, 3, -1, 'fixture' FROM `account` WHERE username = 'ADMIN'
ON DUPLICATE KEY UPDATE gmlevel = VALUES(gmlevel);
INSERT INTO `account_access` (id, gmlevel, RealmID, comment)
SELECT id, 2, 1, 'fixture' FROM `account` WHERE username = 'ADMIN'
ON DUPLICATE KEY UPDATE gmlevel = VALUES(gmlevel);

-- A permanent ban for BANNED (unbandate = 0 means permanent).
INSERT INTO `account_banned` (id, bandate, unbandate, bannedby, banreason, active)
SELECT id, UNIX_TIMESTAMP(), 0, 'fixture', 'fixture ban', 1 FROM `account` WHERE username = 'BANNED'
ON DUPLICATE KEY UPDATE active = VALUES(active), banreason = VALUES(banreason);
