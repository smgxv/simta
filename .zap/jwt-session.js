function processMessage(helper, msg) {
    var token = helper.getGlobalVariables().get("token");
    if (token && token.length > 0) {
      msg.getRequestHeader().setHeader("Authorization", "Bearer " + token);
    }
  }
  