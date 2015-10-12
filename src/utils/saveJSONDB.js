/**
 * Created by Igor Navrotskyj on 24.09.2015.
 */

'use strict';

var MongoClient = require("mongodb").MongoClient,
    mongoClient = new MongoClient(),
    fs = require('fs')
    ;

mongoClient.connect('mongodb://pre.webitel.com:27017/webitel' ,function(err, db) {
    if (err) {
        return console.error('Connect db error: %s', err.message);
    };
    var collection = db.collection('location');
    //importCollection(collection);
    loadCollection(collection);
});

function importCollection (collection) {
    collection
        .find()
        .toArray(function (err, array) {
            if (err) {
                return console.error('Connect db error: %s', err.message);
            };

            //fs.writeFile( "./location.json", JSON.stringify( array ), "utf8", function () {
            //    process.exit(0)
            //} );
        });
};

function loadCollection(collection) {
    var myJson = require("./location.js");
    myJson.forEach(function (i) {
        console.dir(i.country);
        collection.insert(i, function (err) {
            if (err) {
                console.error(err);
            }
        })
    });
};