'use strict';

// 서버세팅
var express = require('express');
var path = require('path');
var fs = require('fs');

// fabric 연결설정
const { Gateway, Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');

const { buildCAClient, registerAndEnrollUser, enrollAdmin } = require('./js/CAUtil.js');
const { buildCCPOrg1, buildWallet } = require('./js/AppUtil.js');

const mspOrg1 = 'Org1MSP';
const walletPath = path.join(__dirname, 'wallet');

const ccp = buildCCPOrg1();
const caClient = buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');

// 미들웨어 설정
var app = express();

// static /public -> ./public 설정
app.use('/public', express.static(path.join(__dirname,'public')));

// body-parser 설정
app.use(express.urlencoded({ extended : false}));
app.use(express.json());

// view 설정
app.set('view engine', 'ejs');
app.set('views', './views');

// '/' GET 라우팅
app.get('/', async(req, res)=>{
    res.sendFile(__dirname + '/public/index.html')
})

// '/user' POST 라우팅
app.post('/user', async(req, res)=>{
    var id = req.body.id;
    var userrole = req.body.userrole;

    console.log("/user start -- ", id, userrole);

    try {
        const wallet = await buildWallet(Wallets, walletPath);

        const userIdentity = await wallet.get(id);
        if (userIdentity) {
            console.log(`An identity for the user ${id} already exists in the wallet`);
            var result = `{"result":"fail", "msg":"An identity for the user ${id} already exists in the wallet"}`;
            var obj = JSON.parse(result);
            console.log("/user end -- failed");
            res.status(200).send(obj);
            return;
        }

        const adminIdentity = await wallet.get('admin');
        if (!adminIdentity) {
            console.log('An identity for the admin user "admin" does not exist in the wallet');
            var result = `{"result":"fail", "msg":"An identity for the admin user admin does not exist in the wallet"}`;
            var obj = JSON.parse(result);
            console.log("/user end -- failed");
            res.status(200).send(obj);
            return;
        }

        // build a user object for authenticating with the CA
        const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
        const adminUser = await provider.getUserContext(adminIdentity, 'admin');

        // Register the user, enroll the user, and import the new identity into the wallet.
        const secret = await caClient.register({
            affiliation: 'org1.department1',
            enrollmentID: id,
            role: userrole
        }, adminUser);
        const enrollment = await caClient.enroll({
            enrollmentID: id,
            enrollmentSecret: secret
        });
        const x509Identity = {
            credentials: {
            certificate: enrollment.certificate,
            privateKey: enrollment.key.toBytes(),
            },
            mspId: 'Org1MSP',
            type: 'X.509',
        };

        await wallet.put(id, x509Identity);
        console.log('Successfully registered and enrolled admin user "appUser" and imported it into the wallet');

    } catch (error) {

    var result = `{"result":"fail", "msg":"Error occured in try/catch in registering userid : ${id}"}`;
    var obj = JSON.parse(result);
        console.log(`/user end -- failed : ${error}`);
        res.status(200).send(obj);
        return;
    }

    var result = `{"result":"success", "msg":"Successfully registered and enrolled admin user ${id} and imported it into the wallet"}`;
    var obj = JSON.parse(result);
    console.log("/user end -- success");
    res.status(200).send(obj);
})

// '/owner' POST 라우팅
app.post('/owner', async(req, res)=>{
    const id = req.body.adminid;
    const adminpw = req.body.passwd;
  
    console.log(id, adminpw);
  
    try {
  
          const wallet = await buildWallet(Wallets, walletPath);
  
      const identity = await wallet.get(id);
      if (identity) {
        console.log('An identity for the admin user admin already exists in the wallet');
        var result = '{"result":"failed", "msg":"An identity for the admin user admin already exists in the wallet"}'
        res.json(JSON.parse(result));
        return;
      }
      // 5. admin등록
      const enrollment = await caClient.enroll({ enrollmentID: id, enrollmentSecret: adminpw });
      const x509Identity = {
      credentials: {
        certificate: enrollment.certificate,
        privateKey: enrollment.key.toBytes(),
        },
        mspId: 'Org1MSP',
        type: 'X.509',
      };
      await wallet.put(id, x509Identity);
  
      console.log('Successfully enrolled admin user and imported it into the wallet');
      var result = '{"result":"success", "msg":"Successfully enrolled admin user and imported it into the wallet"}'
      res.status(200).json(JSON.parse(result));
  
    } catch (error) {
      console.error(`Failed to enroll admin user : ${error}`);
      var result = '{"result":"failed", "msg":"Failed to enroll admin user in try/catch"}'
      res.json(JSON.parse(result));
    }
})

// '/order' POST 라우팅
app.post('/order', async(req,res)=>{
    var name = req.body.name;

    console.log("/order post start -- ",name);
    res.render('order_form.ejs')
})

// url : /asset/tx, method : POST  라우팅 처리 
app.post('/asset/tx', async(req, res) =>{
    // web client요청 문서에서 필요 파라미터들 꺼내오기
    var userid = req.body.userid;
    var fkey = req.body.fromkey;
    var tkey = req.body.tokey;
    var value = req.body.value;

    console.log("/asset/tx post start -- ",userid, fkey, tkey, value);
    const gateway = new Gateway();

    try {
        const wallet = await buildWallet(Wallets, walletPath);

        await gateway.connect(ccp, {
            wallet,
            identity: userid,
            discovery: { enabled: true, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
        });
        
        const network = await gateway.getNetwork("mychannel");
        const contract = network.getContract("simpleasset");
        await contract.submitTransaction('Transfer', fkey, tkey, value);

    } catch (error) {
        var result = `{"result":"fail", "message":"tx has NOT submitted"}`;
        var obj = JSON.parse(result);
        console.log("/asset end -- failed ", error);
        res.status(200).send(obj);
        return;
    }finally {
        gateway.disconnect();
    }

    var result = `{"result":"success", "message":"tx has submitted"}`;
    var obj = JSON.parse(result);
    console.log("/aset end -- success");
    res.status(200).send(obj);
});

// 6. 서버 listen (서버시작)
// server listen
app.listen(3000, () => {
    console.log('Express server is started: 3000');
});