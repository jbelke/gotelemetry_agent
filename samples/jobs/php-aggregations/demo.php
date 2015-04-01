#!/usr/bin/env php
<?php

echo '{"$push":{"$series":"test","$value":123}}';
echo "\n";
echo '{"value":{"$add": {"$left": 10, "$right": 20}}}';
