/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric_one` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag_one` text DEFAULT NULL,
  `tag_two` text DEFAULT NULL,
  `int64_one` int(11) DEFAULT NULL,
  `int64_two` int(11) DEFAULT NULL,
  `bool_one` tinyint(1) DEFAULT NULL,
  `bool_two` tinyint(1) DEFAULT NULL,
  `uint64_one` int(10) unsigned DEFAULT NULL,
  `float64_one` double DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric_one` VALUES ('2021-05-17 22:04:45','tag1','tag2',1234,2345,1,0,1000000000,3.1415);
