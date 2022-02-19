const config = require('config');
const crypto = require('crypto');
const express = require('express');
const request = require('request');

const host = config.get('server.host');
const port = config.get('server.port');
const be_host = config.get('be.host');
const be_port = config.get('be.port');

const algorithm = config.get('crypto.algorithm');
const cipher_key = config.get('crypto.cipher_key');
const block_size = config.get('crypto.block_size');

const app = express();

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
            '/get_users',
            '/get_users_encr',
            '/get_users_decr',
        ],
    });
});

app.get('/get_users', (req, res) => {
    let url = `http://${be_host}:${be_port}/get_users`;
    request(url, function (error, response, body) {        
        let data = JSON.parse(body)
        
        if (!error && response.statusCode == 200) {
            res.json(data)
        }
    })
});

app.get('/get_users_decr', (req, res) => {
    let url = `http://${be_host}:${be_port}/get_users_encr`;
    request(url, function (error, response, body) {        
        let data = JSON.parse(body)

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