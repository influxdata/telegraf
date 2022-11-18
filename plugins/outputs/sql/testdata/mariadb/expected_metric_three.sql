/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric three` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag four` text DEFAULT NULL,
  `string two` text DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric three` VALUES ('2021-05-17 22:04:45','tag4','string2');
