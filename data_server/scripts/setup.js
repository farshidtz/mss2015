(function () {
  'use strict';

  var console = require("console"),
  		db = require("org/arangodb").db,
  		_ = require('underscore');

  var cfg = applicationContext.configuration;
  		
  var createContextCollection = function(collection) {
    var name = applicationContext.collectionName(collection);
    if (db._collection(name) === null) {
      db._create(name);
    } else if (applicationContext.isProduction) {
      console.warn("collection '%s' already exists. Leaving it untouched.", name);
    }
  };
	
	var createGenericCollection = function(collection) {
    if (db._collection(collection) === null) {
      db._create(collection);
    } else if (applicationContext.isProduction) {
      console.warn("collection '%s' already exists. Leaving it untouched.", collection);
    }
  };
  
  var collections = [cfg.collectionName];
  
  //createContextCollection();
	_.forEach(collections, function(c){
		createGenericCollection(c);
	});
  
}());