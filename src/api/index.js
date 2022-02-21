const config = require('config');
const crypto = require('crypto');
const express = require('express');
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
            '/get_employees',
            // '/get_employees_encr',
            '/get_employees_decr',
        ],
    });
});

app.get('/get_employees', (req, res) => {
    let url = `http://${be_host}:${be_port}/get_employees`;

    request(url, function (error, response, body) {        
        let data = JSON.parse(body)
        
        if (!error && response.statusCode == 200) {
            res.json(data)
        }
    })
});

app.get('/get_employees_decr', (req, res) => {
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
    let url = `http://${be_host}:${be_port}/get_employees_encr?`;
    
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

        if (!error && response.statusCode == 200) {
            res.json(data)
        }
    })
});

app.listen(port, () => console.log(`api 1 listening on port ${port}!`))