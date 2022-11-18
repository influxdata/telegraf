/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric_two` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag_three` text DEFAULT NULL,
  `string_one` text DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric_two` VALUES ('2021-05-17 22:04:45','tag3','string1');
