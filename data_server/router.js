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
    * List data: by sensor and position
    *
    */
    app.get("/list", function(req, res){
    	//sensor name
    	let sensor_name = req.params("n");
        let context_name = req.params("p");
		let start = req.params("start");
		let end = req.params("end");

    	//Retrive data from collection
    	//let data = col.byExample({"n" : sensor_name}).toArray();
    	let query = Foxx.createQuery({
    		query: 'FOR n IN @@collectionName FILTER n.n == @sensorName AND n.p == @contextName AND DATE_ISO8601(n.t) > @startTime AND DATE_ISO8601(n.t) < @endTime SORT n.t ASC RETURN {"t" : n.t, "time" : DATE_ISO8601(n.t), "v0" : n.v0, "v1" : n.v1, "v2" : n.v2, "v3" : n.v3, "v4" : n.v4, "v5" : n.v5, "n": n.n, "p": n.p, "e": n.e}'
    	});

    	let bind = {
    		'@collectionName': cfg.collectionName,
    		sensorName : sensor_name,
            contextName : context_name,
			startTime : start,
			endTime : end,
    	};

    	res.json(query(bind));

    }).queryParam("n", joi.string().insensitive().required()).queryParam("p", joi.string().insensitive().required()).queryParam("start", joi.string().insensitive().required()).queryParam("end", joi.string().insensitive().required());
	
	
    /**
    *
    * List data: by sensor
    *
    */	
	app.get("/list2", function(req, res){
    	//sensor name
    	let sensor_name = req.params("n");
		let start = req.params("start");
		let end = req.params("end");

    	//Retrive data from collection
		let query = Foxx.createQuery({
    		query: 'FOR n IN @@collectionName FILTER n.n == @sensorName AND DATE_ISO8601(n.t) > @startTime AND DATE_ISO8601(n.t) < @endTime SORT n.t ASC RETURN {"t" : n.t, "v0" : n.v0, "v1" : n.v1, "v2" : n.v2, "v3" : n.v3, "v4" : n.v4, "v5" : n.v5, "n": n.n, "p": n.p, "e": n.e}'
    	});

    	let bind = {
    		'@collectionName': cfg.collectionName,
    		sensorName : sensor_name,
			startTime : start,
			endTime : end
    	};

    	res.json(query(bind));

	}).queryParam("n", joi.string().insensitive().required()).queryParam("start", joi.string().insensitive().required()).queryParam("end", joi.string().insensitive().required());
	
	/**
    *
    * List data: by position
    *
    */	
	app.get("/list3", function(req, res){
    	//sensor name
    	let context_name = req.params("p");
		let start = req.params("start");
		let end = req.params("end");

    	//Retrive data from collection
		let query = Foxx.createQuery({
    		query: 'FOR n IN @@collectionName FILTER n.p == @contextName AND DATE_ISO8601(n.t) > @startTime AND DATE_ISO8601(n.t) < @endTime SORT n.t ASC RETURN {"t" : n.t, "v0" : n.v0, "v1" : n.v1, "v2" : n.v2, "v3" : n.v3, "v4" : n.v4, "v5" : n.v5, "n": n.n, "p": n.p, "e": n.e}'
    	});

    	let bind = {
    		'@collectionName': cfg.collectionName,
    		contextName : context_name,
			startTime : start,
			endTime : end
    	};

    	res.json(query(bind));

	}).queryParam("p", joi.string().insensitive().required()).queryParam("start", joi.string().insensitive().required()).queryParam("end", joi.string().insensitive().required());

})();