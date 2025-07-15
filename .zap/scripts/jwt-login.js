function authenticate(helper, paramsValues, credentials) {
    var email = credentials.getParam("email");
    var password = credentials.getParam("password");
  
    var loginMsg = helper.prepareMessage();
    loginMsg.setRequestHeader("POST /login HTTP/1.1\r\nHost: 104.43.89.154:8080\r\nContent-Type: application/json");
    loginMsg.setRequestBody(JSON.stringify({ email: email, password: password }));
  
    helper.sendAndReceive(loginMsg, false);
    var json = JSON.parse(loginMsg.getResponseBody().toString());
  
    if (json.token) {
      helper.getGlobalVariables().set("token", json.token);
    }
  
    return true;
  }
  