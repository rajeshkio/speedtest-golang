
CREATE DATABASE if not exists speedtest;
use speedtest;

DROP TABLE IF EXISTS `speedtest`;

CREATE TABLE if not exists `speedtest` (
  `id` int NOT NULL AUTO_INCREMENT,
  `TimeStamp` datetime NOT NULL,
  `DownloadSpeed` varchar(50) DEFAULT NULL,
  `UploadSpeed` varchar(50) DEFAULT NULL,
  `Latency` varchar(50) DEFAULT NULL,
  `PublicIp` varchar(250) DEFAULT NULL,
  `ISP` varchar(250) DEFAULT NULL,
  `Peers` varchar(250) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;

