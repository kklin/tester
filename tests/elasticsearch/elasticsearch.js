const {createDeployment} = require("@quilt/quilt");
var Elasticsearch = require("@quilt/elasticsearch").Elasticsearch;
var infrastructure = require("@quilt/tester/config/infrastructure")

var deployment = createDeployment({});
deployment.deploy(infrastructure);
deployment.deploy(new Elasticsearch(infrastructure.nWorker).public());
