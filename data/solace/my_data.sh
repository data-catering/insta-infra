#!/bin/sh

host=solace

#Swagger doc: https://docs.solace.com/API-Developer-Online-Ref-Documentation/swagger-ui/software-broker/config/index.html
echo "Creating new queue in Solace"
curl "http://$host:8080/SEMP/v2/config/msgVpns/default/queues" \
  -X POST \
  -u admin:admin \
  -H "Content-type:application/json" \
  -d '{ "queueName":"rest_test_queue","accessType":"exclusive","maxMsgSpoolUsage":200,"permission":"consume","ingressEnabled":true,"egressEnabled":true }'

echo "Creating JNDI queue object"
curl "http://$host:8080/SEMP/v2/config/msgVpns/default/jndiQueues" \
  -X POST \
  -u admin:admin \
  -H "Content-type:application/json" \
  -d '{ "msgVpnName":"default","physicalName":"rest_test_queue","queueName":"/JNDI/Q/rest_test_queue" }'

echo "Creating new topic in Solace"
curl http://$host:8080/SEMP/v2/config/msgVpns/default/topicEndpoints \
  -X POST \
  -u admin:admin \
  -H "Content-type:application/json" \
  -d '{ "topicEndpointName":"rest_test_topic","accessType":"exclusive","maxSpoolUsage":200,"permission":"consume","ingressEnabled":true,"egressEnabled":true }'

echo "Creating JNDI queue object"
curl http://$host:8080/SEMP/v2/config/msgVpns/default/jndiTopics \
  -X POST \
  -u admin:admin \
  -H "Content-type:application/json" \
  -d '{ "msgVpnName":"default","physicalName":"rest_test_topic","topicName":"/JNDI/T/rest_test_topic" }'

echo "Creating subscription between queue and topic"
curl http://$host:8080/SEMP/v2/config/msgVpns/default/queues/rest_test_queue/subscriptions \
  -X POST \
  -u admin:admin \
  -H "Content-type:application/json" \
  -d '{ "msgVpnName":"default","queueName":"rest_test_queue","subscriptionTopic":"rest_test_topic" }'

