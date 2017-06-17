const {createDeployment} = require("@quilt/quilt");
var infrastructure = require("@quilt/tester/config/infrastructure.js")

createDeployment({}).deploy(infrastructure);
