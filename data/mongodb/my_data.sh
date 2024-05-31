#!/usr/bin/env bash

sleep 5
echo "
db = connect( 'mongodb://root:root@mongodb' );

db.movies.insertMany( [
   {
      title: 'Titanic',
      year: 1997,
      genres: [ 'Drama', 'Romance' ]
   },
   {
      title: 'Spirited Away',
      year: 2001,
      genres: [ 'Animation', 'Adventure', 'Family' ]
   },
   {
      title: 'Casablanca',
      genres: [ 'Drama', 'Romance', 'War' ]
   }
] );
" > /tmp/mongo_startup.js

mongosh "mongodb://root:root@mongodb" -f /tmp/mongo_startup.js

mongosh "mongodb://root:root@mongodb" --eval "show tables; db.movies.findOne();"

tail -f /dev/null
