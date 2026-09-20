-- AzerothCore world database (acore_world) fixture for local development.
--
-- The real world DB is huge and mostly game content, so this fixture keeps only
-- the columns the gateway's item catalog and render read from `item_template`.
-- The adapter selects only real AzerothCore columns, so it runs unchanged
-- against a full world database.

CREATE DATABASE IF NOT EXISTS acore_world;
-- The gateway only reads from AzerothCore databases; grant SELECT only.
GRANT SELECT ON acore_world.* TO 'acgw'@'%';
FLUSH PRIVILEGES;

USE acore_world;

DROP TABLE IF EXISTS `item_template`;
CREATE TABLE `item_template` (
  `entry`          mediumint unsigned NOT NULL DEFAULT 0,
  `class`          tinyint unsigned NOT NULL DEFAULT 0,
  `subclass`       tinyint unsigned NOT NULL DEFAULT 0,
  `name`           varchar(255) NOT NULL,
  `displayid`      mediumint unsigned NOT NULL DEFAULT 0,
  `Quality`        tinyint unsigned NOT NULL DEFAULT 0,
  `Flags`          bigint NOT NULL DEFAULT 0,
  `BuyCount`       tinyint unsigned NOT NULL DEFAULT 1,
  `BuyPrice`       bigint NOT NULL DEFAULT 0,
  `SellPrice`      int unsigned NOT NULL DEFAULT 0,
  `InventoryType`  tinyint unsigned NOT NULL DEFAULT 0,
  `AllowableClass` int NOT NULL DEFAULT -1,
  `AllowableRace`  int NOT NULL DEFAULT -1,
  `ItemLevel`      smallint unsigned NOT NULL DEFAULT 0,
  `RequiredLevel`  tinyint unsigned NOT NULL DEFAULT 0,
  `maxcount`       int NOT NULL DEFAULT 0,
  `stackable`      int NOT NULL DEFAULT 1,
  `ContainerSlots` smallint unsigned NOT NULL DEFAULT 0,
  `bonding`        tinyint unsigned NOT NULL DEFAULT 0,
  `description`    varchar(255) NOT NULL DEFAULT '',
  `armor`          smallint unsigned NOT NULL DEFAULT 0,
  `holy_res`       tinyint unsigned NOT NULL DEFAULT 0,
  `fire_res`       tinyint unsigned NOT NULL DEFAULT 0,
  `nature_res`     tinyint unsigned NOT NULL DEFAULT 0,
  `frost_res`      tinyint unsigned NOT NULL DEFAULT 0,
  `shadow_res`     tinyint unsigned NOT NULL DEFAULT 0,
  `arcane_res`     tinyint unsigned NOT NULL DEFAULT 0,
  `Delay`          smallint unsigned NOT NULL DEFAULT 1000,
  `itemset`        mediumint unsigned NOT NULL DEFAULT 0,
  `stat_type1`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value1`    smallint NOT NULL DEFAULT 0,
  `stat_type2`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value2`    smallint NOT NULL DEFAULT 0,
  `stat_type3`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value3`    smallint NOT NULL DEFAULT 0,
  `stat_type4`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value4`    smallint NOT NULL DEFAULT 0,
  `stat_type5`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value5`    smallint NOT NULL DEFAULT 0,
  `stat_type6`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value6`    smallint NOT NULL DEFAULT 0,
  `stat_type7`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value7`    smallint NOT NULL DEFAULT 0,
  `stat_type8`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value8`    smallint NOT NULL DEFAULT 0,
  `stat_type9`     tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value9`    smallint NOT NULL DEFAULT 0,
  `stat_type10`    tinyint unsigned NOT NULL DEFAULT 0,
  `stat_value10`   smallint NOT NULL DEFAULT 0,
  `dmg_min1`       float NOT NULL DEFAULT 0,
  `dmg_max1`       float NOT NULL DEFAULT 0,
  `dmg_type1`      tinyint unsigned NOT NULL DEFAULT 0,
  `dmg_min2`       float NOT NULL DEFAULT 0,
  `dmg_max2`       float NOT NULL DEFAULT 0,
  `dmg_type2`      tinyint unsigned NOT NULL DEFAULT 0,
  `spellid_1`      mediumint NOT NULL DEFAULT 0,
  `spelltrigger_1` tinyint unsigned NOT NULL DEFAULT 0,
  `spellcooldown_1` int NOT NULL DEFAULT -1,
  `spellid_2`      mediumint NOT NULL DEFAULT 0,
  `spelltrigger_2` tinyint unsigned NOT NULL DEFAULT 0,
  `spellcooldown_2` int NOT NULL DEFAULT -1,
  `spellid_3`      mediumint NOT NULL DEFAULT 0,
  `spelltrigger_3` tinyint unsigned NOT NULL DEFAULT 0,
  `spellcooldown_3` int NOT NULL DEFAULT -1,
  `spellid_4`      mediumint NOT NULL DEFAULT 0,
  `spelltrigger_4` tinyint unsigned NOT NULL DEFAULT 0,
  `spellcooldown_4` int NOT NULL DEFAULT -1,
  `spellid_5`      mediumint NOT NULL DEFAULT 0,
  `spelltrigger_5` tinyint unsigned NOT NULL DEFAULT 0,
  `spellcooldown_5` int NOT NULL DEFAULT -1,
  PRIMARY KEY (`entry`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `item_template`
  (entry, class, subclass, name, displayid, Quality, BuyPrice, SellPrice, InventoryType,
   ItemLevel, RequiredLevel, maxcount, stackable, ContainerSlots, bonding, description,
   armor, Delay, stat_type1, stat_value1, stat_type2, stat_value2,
   dmg_min1, dmg_max1, dmg_type1, spellid_1, spelltrigger_1, spellcooldown_1)
VALUES
  (4496,  1,  0, 'Traveler''s Backpack', 2448, 1, 0, 2500, 18,
   1, 0, 0, 1, 16, 0, 'A 16-slot bag.',
   0, 1000, 0, 0, 0, 0,
   0, 0, 0, 0, 0, -1),
  (6948,  15, 0, 'Hearthstone', 64178, 1, 0, 0, 0,
   1, 0, 1, 1, 0, 1, 'Returns you to your home location.',
   0, 1000, 0, 0, 0, 0,
   0, 0, 0, 0, 0, -1),
  (19019, 2,  7, 'Thunderfury, Blessed Blade of the Windseeker', 30606, 5, 0, 0, 13,
   80, 60, 1, 1, 0, 1, 'A legendary blade.',
   0, 1900, 0, 0, 0, 0,
   44, 115, 0, 21992, 2, -1),
  (200000, 2, 7, 'Community Blade', 30606, 4, 0, 0, 13,
   70, 60, 1, 1, 0, 1, 'A reward from the community store.',
   0, 2000, 4, 20, 7, 15,
   50, 90, 0, 0, 0, -1)
ON DUPLICATE KEY UPDATE name = VALUES(name), Quality = VALUES(Quality);
