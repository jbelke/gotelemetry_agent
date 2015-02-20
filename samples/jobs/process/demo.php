#!/usr/bin/env php
<?php

// echo "REPLACE\n";
// echo '{"value":' . $argv[2] . '}';
echo "PATCH\n";
echo '[{"op":"replace", "path":"/value", "value":' . $argv[2] . '}]';