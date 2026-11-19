CREATE UNIQUE INDEX kafka_outbox_topic_key ON kafka_outbox (topic, message_key);
