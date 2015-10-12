/**
 * Created by Igor Navrotskyj on 22.09.2015.
 */

'use strict';

module.exports = {
    addQuery: addQuery
};

function addQuery (db) {
    return {
        find_: function(number, cb) {
            var collection = db.collection('locationTEST');

            var numbers = [];
            number.split('').reduce(function(r, v, i, a) {
                numbers.push(a.slice(0, i).join(''));
                return r + v
            });

            console.log(numbers);

            // Map function
            var map = 'function() { \n' +
                '   if (this.sysSearch.test("' + number +'")) \n' +
                '       emit(this.country, this); \n' +
            '}';

            // Reduce function
            var reduce = 'function(k, values) { \n' +
                '   var result = values[0]; \n' +
                '   values.forEach(function(value) { \n' +
                '       if (value.sysOrder > result.sysOrder) \n' +
                '           result = value; \n' +
                '   }); \n' +
                '   return result;\n' +
                '}';
            // Peform the map reduce
            collection.mapReduce(map, reduce, {
                out: {inline : 1}
            }, function (err, res) {
                if (err)
                    throw 1;
                cb(err, res);
            });
        },
        
        find: function (number, cb) {
            var collection = db.collection('locationTEST');

            var numbers = [];
            number.split('').reduce(function(r, v, i, a) {
                numbers.push(a.slice(0, i).join(''));
                return r + v
            });

            collection
                .find({"sysLength": number.length,"code2": { $in: numbers}})
                .sort({"sysOrder": -1})
                .limit(1)
                .toArray(cb);
        }
    }
}