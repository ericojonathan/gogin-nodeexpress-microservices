const config = require('config');
const crypto = require('crypto');
const express = require('express');
var http = require('http');
const request = require('request');
const querystring = require('querystring');
const url = require('url');
const { query } = require('express');

const api_key = config.get('app.api_key');
const host = config.get('server.host');
const port = config.get('server.port');
const be_host = config.get('be.host');
const be_port = config.get('be.port');

const algorithm = config.get('crypto.algorithm');
const cipher_key = config.get('crypto.cipher_key');
const block_size = config.get('crypto.block_size');

const app = express();
app.use(express.json());

function encrypt(plainText) {
    const iv = crypto.randomBytes(block_size);
    const cipher = crypto.createCipheriv(algorithm, cipher_key, iv);
    let cipherText;
    try {
      cipherText = cipher.update(plainText, 'utf8', 'hex');
      cipherText += cipher.final('hex');
      cipherText = iv.toString('hex') + cipherText
    } catch (e) {
      cipherText = null;
    }
    return cipherText;
  }

function decrypt(cipherText) {
    const contents = Buffer.from(cipherText, 'hex');
    const iv = contents.slice(0, block_size);
    const textBytes = contents.slice(block_size);
  
    const decipher = crypto.createDecipheriv(algorithm, cipher_key, iv);
    let decrypted = decipher.update(textBytes, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
  }

app.get('/', (req, res) => {
    res.json({
        name: 'api 1',
        services: [
            '/employees',
            // '/get_employees_encr',
            '/employees_unecr',
        ],
    });
});

app.get('/employees_unencr', (req, res) => {
    let url = `http://${be_host}:${be_port}/employees_unencr`;

    request(url, function (error, response, body) {        
        let data = JSON.parse(body)
        
        if (!error && response.statusCode == 200) {
            res.json(data)
        }
    })
});

app.delete('/employees', (req, res) => {
    //cache will be based on an API Key. 
    //access without API key will be rejected
    let key_recv = req.query.api_key;    
    console.log("key: " + key_recv);    
    if(key_recv === undefined || key_recv === null || key_recv =='' || key_recv != api_key) {
        res.status(401).send("Unauthorized access! Please use your API Key to use this service.")
        return;
    }
    let emp_id = req.query.emp_id;    
    if(emp_id === undefined || emp_id == '') {
        res.status(400).send("Bad Request!")
        return;
    }

    var data = JSON.stringify({
        id: emp_id        
    });

    var options = {
        host: 'localhost',
        port: 3000,
        path: '/employees_encr',
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Content-Length': Buffer.byteLength(data)
        }
    };

    var httpreq = http.request(options, function (response) {
        response.setEncoding('utf8');
        response.on('data', function (chunk) {
          console.log(chunk);
        });
        response.on('end', function() {
          res.send('ok');

        });        
    });
    
    httpreq.on('error', function (e) {
        res.status(400).send("Bad Request!")
        return;
    });

    httpreq.on('timeout', function(e){
        res.status(400).send("Bad Request!")
        return;
    });

    httpreq.setTimeout(5000);
    httpreq.write(data);
    httpreq.end();     

    console.log("deleting id: " + emp_id);
    // res.json({"message": "data deleted!"});
    return;
});

app.put('/employees', (req, res) => {
    //cache will be based on an API Key. 
    //access without API key will be rejected
    console.log("[UPDATING EMPLOYEES]")
    let key_recv = req.body.api_key;
    console.log("key_recv: " + key_recv);
    if(key_recv === undefined || key_recv === null || key_recv =='' || key_recv != api_key) {
        res.status(401).send("Unauthorized access! Please use your API Key to use this service.")
        return;
    }

    let emp_id = req.body.emp_id;
    let job_title = req.body.job_title;
    let email_address = req.body.email_address;
    let firstName_LastName = req.body.firstName_LastName;
    
    if(emp_id === undefined || job_title === undefined || email_address === undefined || firstName_LastName === undefined ) {
        console.log("ERROR: undefined");
        res.status(400).send("<h1>Update error!</h1><p>E.g. job_title, email_address and firstName_LastNames fields are required.</p>")
        return;
    }

    //email address format
    //other checks, e.g. acceptable job titles, etc.
    emp_id = encrypt(emp_id)
    job_title = encrypt(job_title);
    email_address = encrypt(email_address);
    firstName_LastName = encrypt(firstName_LastName);
    
    var data = JSON.stringify({
        id: emp_id,
        job_title: job_title,
        email_address: email_address,
        firstName_LastName: firstName_LastName
    });

    var options = {
        host: 'localhost',
        port: 3000,
        path: '/employees_encr',
        method: 'PUT',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Content-Length': Buffer.byteLength(data)
        }
    };

    var httpreq = http.request(options, function (response) {
        response.setEncoding('utf8');
        response.on('data', function (chunk) {
          console.log(chunk);
        });
        response.on('end', function() {
          res.send('ok');
        })
    });
    
    httpreq.write(data);
    httpreq.end();     
});

app.post('/employees', (req, res) => {
    //cache will be based on an API Key. 
    //access without API key will be rejected
    let key_recv = req.body.api_key;
    if(key_recv === undefined || key_recv === null || key_recv =='' || key_recv != api_key) {
        res.status(401).send("Unauthorized access! Please use your API Key to use this service.")
        return;
    }

    let job_title = req.body.job_title;
    let email_address = req.body.email_address;
    let firstName_LastName = req.body.firstName_LastName;

    if(job_title === undefined || email_address === undefined || firstName_LastName === undefined ) {
        res.status(400).send("<h1>Post error!</h1><p>E.g. job_title, email_address and firstName_LastNames fields are required.</p>")
        return;
    }

    //email address format
    //other checks, e.g. acceptable job titles, etc.

    job_title = encrypt(job_title);
    email_address = encrypt(email_address);
    firstName_LastName = encrypt(firstName_LastName);
    
    var data = JSON.stringify({
        job_title: job_title,
        email_address: email_address,
        firstName_LastName: firstName_LastName
    });

    var options = {
        host: 'localhost',
        port: 3000,
        path: '/employees_encr',
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
          'Content-Length': Buffer.byteLength(data)
        }
    };

    var httpreq = http.request(options, function (response) {
        response.setEncoding('utf8');
        response.on('data', function (chunk) {
          console.log(chunk);
        });
        response.on('end', function() {
          res.send('ok');
        })
    });
    
    httpreq.write(data);
    httpreq.end();     
});

app.get('/employees', (req, res) => {
    //cache will be based on an API Key. 
    //access without API key will be rejected
    let key_recv = req.query.api_key;
    if(key_recv === undefined || key_recv === null || key_recv =='' || key_recv != api_key) {
        res.status(401).send("Unauthorized access! Please use your API Key to use this service.")
        return;
    }

    let query_length = Object.keys(req.query).length;
    if(query_length > 2) {
        res.status(400).send("<h1>Query error!</h1><p>For now, query is based on <b>one</b> key=value only.</p><p>E.g. job_title=Manager. job_title=Manager&email=example@example.com will produce this query error.</p>")
        return;
    }

    let key_encr = encrypt(api_key);
    let url = `http://${be_host}:${be_port}/employees_encr?`;
    
    if(query_length > 1) {
        //iterates
        for(key in req.query) {
            if(key =="api_key") {                                
                continue;
            }
            url += `${key}=${req.query[key]}&`;
        }        
    } 
    url += `api_key=${key_encr}`;
    request(url, function (error, response, body) {
        
        if(error !== null) {
            console.log("error.status: " + error.status)
            res.status(error.status || 500).send({
                error: {
                  status: error.status || 500,
                  message: 'Service Error',
                },
            });
            return;
        } else if(error === null && body.trim() =='') {
            res.status(404).send({
                error: {
                  status: 404,
                  message: 'Not Found',
                },
            });
            return;
        }
                
        let data = JSON.parse(body);    
        
        for(let i = 0; i < data.length; i++) {
            let obj = data[i];
            for(let key in obj) {
                if(key != "id") {
                    obj[key] = decrypt(obj[key]);
                }
            }
        }

        res.json(data)        
    })
});

//Inter-service communication using sync and async communication
app.get('/employees_async', async (req, res) => {
    //cache will be based on an API Key. 
    //access without API key will be rejected
    let key_recv_primary = req.query.api_key_primary;
    if(key_recv_primary === undefined || key_recv_primary === null || key_recv_primary =='' || key_recv_primary != api_key) {
        res.status(401).send("Unauthorized access! Please use your API Key to use this service.")
        return;
    }

    let api_key_async = config.get('async.api_key');
    let key_recv_async = req.query.api_key_secondary;
    if(key_recv_async === undefined || key_recv_async === null || key_recv_async =='' || key_recv_async != api_key_async) {
        res.status(401).send("Unauthorized access! Please use your secondary API Key to use this async service.")
        return;
    }

    let query_length = Object.keys(req.query).length;
    if(query_length > 2) {
        res.status(400).send("<h1>Query error!</h1><p>For now, query is based on <b>one</b> key=value only.</p><p>E.g. job_title=Manager. job_title=Manager&email=example@example.com will produce this query error.</p>")
        return;
    }

    //Values from config file
    let async_host = config.get("async.host");
    let async_port = config.get("async.port");
    let url = `http://${async_host}:${async_port}/employees?`;

    if(query_length > 1) {
        //iterates
        for(key in req.query) {
            if(key =="api_key") {                                
                continue;
            }
            url += `${key}=${req.query[key]}&`;
        }        
    }

    url += `api_key=${api_key_async}`;
    request(url, function (error, response, body) {
        
        if(error !== null) {
            console.log("error.status: " + error.status)
            res.status(error.status || 500).send({
                error: {
                  status: error.status || 500,
                  message: 'Service Error',
                },
            });
            return;
        } else if(error === null && body.trim() =='') {
            res.status(404).send({
                error: {
                  status: 404,
                  message: 'Not Found',
                },
            });
            return;
        }
                
        let data = JSON.parse(body);    
        
        // for(let i = 0; i < data.length; i++) {
        //     let obj = data[i];
        //     for(let key in obj) {
        //         if(key != "id") {
        //             obj[key] = decrypt(obj[key]);
        //         }
        //     }
        // }

        res.json(data)        
    })
})

app.listen(port, () => console.log(`API_2 listening on port ${port}!`))