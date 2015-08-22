(function() {
    'use strict';

    /*
     *
     * Dependencies
     *
     */
    const Foxx = require('org/arangodb/foxx');
    const app = new Foxx.Controller(applicationContext);
    const console = require('console');
    const _ = require('underscore');
    const joi = require('joi');
    const db = require("org/arangodb").db;
    const cfg = applicationContext.configuration;
    const col = db._collection(cfg.collectionName);



    /*
    *
    * Errors
    *
    */
    let BadRequest = function(msg){
    	this.message = msg;
    };
    BadRequest.prototype = new Error();
    


   /*==========  REST Endpoints  ==========*/
   
   /**
   *
   * Save data
   *
   */
    app.post("/save", function(req, res) {
        let body = req.body();
        if (body.Entries) {
            _.each(body.Entries, function(elem, index) {
                //append each record to the collection
                col.insert(elem);
            });

            res.json({
                code: 200
            });
        } else {
        	//throw an error
        	throw new BadRequest("Please provide the data under Entries attribute");
        }

    }).errorResponse(BadRequest, 400, 'Bad Request', function(e){
    	return {
    		error: true,
    		code: 400,
    		errorMessage : e.message
    	};
    });



    /**
    *
    * List data
    *
    */
    app.get("/list", function(req, res){
    	//sensor name
    	let sensor_name = req.params("sensor");
        let context_name = req.params("context");

    	//Retrive data from collection
    	//let data = col.byExample({"n" : sensor_name}).toArray();
    	let query = Foxx.createQuery({
    		query: 'FOR n IN @@collectionName FILTER n.n == @sensorName AND n.p == @contextName RETURN {"time" : DATE_ISO8601(n.t), "v0" : n.v0, "v1" : n.v1, "v2" : n.v2, "v3" : n.v3, "v4" : n.v4, "v5" : n.v5, "sensor_name": n.n, "context": n.p}'
    	});

    	let bind = {
    		'@collectionName': cfg.collectionName,
    		sensorName : sensor_name,
            contextName : context_name
    	};

    	res.json(query(bind));

    }).queryParam("sensor", joi.string().insensitive().required()).queryParam("context", joi.string().insensitive().required());

})();