#!/usr/bin/env node

'use strict';

var express = require('express');
var app = express();

app.get('/test', function(req, res) {
	res.status(400).send('Caterpillar drive offline.');
	// res.send('{"value":123}');
});

app.listen(8000);