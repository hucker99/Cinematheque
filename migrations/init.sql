CREATE DATABASE IF NOT EXISTS testdb;
SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `actors`;
CREATE TABLE `actors` (
    `actor_id`  int(11)         NOT NULL AUTO_INCREMENT,
    `name`      varchar(255)    NOT NULL,
    `gender`    varchar(10)     NOT NULL CHECK (gender IN ('male', 'female')),
    `birthday`  varchar(255)    NOT NULL,
    PRIMARY KEY (`actor_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `films`;
CREATE TABLE `films` (
    `film_id`       int(11)         NOT NULL AUTO_INCREMENT,
    `name`          varchar(255)    NOT NULL,
    `release_date`  varchar(255)    NOT NULL,
    `rating`        varchar(255)    NOT NULL,
    PRIMARY KEY (`film_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `FilmMembership`;
CREATE TABLE `FilmMembership` (
    `actor`  int(11)     NOT NULL,
    `film`   int(11)     NOT NULL,
    PRIMARY KEY (`actor`, `film`),
    CONSTRAINT `Constr_FilmMembership_Actor_fk`
        FOREIGN KEY `Actor_fk` (`actor`) REFERENCES `actors`(`actor_id`)
            ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `Constr_FilmMembership_Film_fk`
        FOREIGN KEY `Film_fk` (`film`) REFERENCES `films` (`film_id`)
            ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
                          `id`          int(11)         NOT NULL AUTO_INCREMENT,
                          `email`       varchar(255)    NOT NULL,
                          `password`    varchar(255)    NOT NULL,
                          `role`        varchar(10)     NOT NULL,
                          PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions` (
                         `uid`          int(11)         NOT NULL,
                         `cookie`       varchar(255)    NOT NULL,
                         `expire_date`  varchar(255)    NOT NULL,
                         PRIMARY KEY (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;