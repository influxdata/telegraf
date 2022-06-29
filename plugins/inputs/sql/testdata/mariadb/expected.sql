/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `bar` (
  `baz` int(11) DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `bar` VALUES (1);
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric three` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag four` text DEFAULT NULL,
  `string two` text DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric three` VALUES ('2021-05-17 22:04:45','tag4','string2');
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric_one` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag_one` text DEFAULT NULL,
  `tag_two` text DEFAULT NULL,
  `int64_one` int(11) DEFAULT NULL,
  `int64_two` int(11) DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric_one` VALUES ('2021-05-17 22:04:45','tag1','tag2',1234,2345);
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `metric_two` (
  `timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
  `tag_three` text DEFAULT NULL,
  `string_one` text DEFAULT NULL
);
/*!40101 SET character_set_client = @saved_cs_client */;
INSERT INTO `metric_two` VALUES ('2021-05-17 22:04:45','tag3','string1');
