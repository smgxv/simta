// @name jwt-login
// @scriptType authentication
// @description Script to authenticate and retrieve JWT token
// @engine ECMAScript : Oracle Nashorn

function authenticate(helper, paramsValues, credentials) {
  var email = credentials.getParam("email");
  var password = credentials.getParam("password");

  var loginMsg = helper.prepareMessage();
  loginMsg.setRequestHeader("POST /login HTTP/1.1\r\nHost: 104.43.89.154:8080\r\nContent-Type: application/json");
  loginMsg.setRequestBody(JSON.stringify({ email: email, password: password }));

  helper.sendAndReceive(loginMsg, false);

  var responseBody = loginMsg.getResponseBody().toString();
  var responseStatus = loginMsg.getResponseHeader().getStatusCode();

  if (responseStatus === 200) {
      try {
          var json = JSON.parse(responseBody);
          if (json.token) {
              helper.getGlobalVariables().set("token", "Bearer " + json.token);
              return true;
          }
      } catch (e) {
          print("Login JSON parsing failed: " + e);
      }
  } else {
      print("Login request failed with status: " + responseStatus);
  }

  return false;
}
