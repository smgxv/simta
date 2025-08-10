const BASE_URL   = java.lang.System.getenv("TARGET_URL") || "https://securesimta.my.id";
const LOGIN_PATH = "/login";
const EMAIL      = java.lang.System.getenv("LOGIN_EMAIL") || "daniel@gmail.com";
const PASSWORD   = java.lang.System.getenv("LOGIN_PASSWORD") || "P@ssword1";
const VAR_KEY    = "jwt_taruna_token";

const URI               = Java.type('org.apache.commons.httpclient.URI');
const HttpRequestHeader = Java.type('org.parosproxy.paros.network.HttpRequestHeader');
const HttpMessage       = Java.type('org.parosproxy.paros.network.HttpMessage');
const ScriptVars        = Java.type('org.zaproxy.zap.extension.script.ScriptVars');

function log(m){ print('[JWT][taruna] ' + m); }

function doLogin(helper){
  const loginUrl = BASE_URL + LOGIN_PATH;
  const reqHeader = new HttpRequestHeader(
    'POST ' + LOGIN_PATH + ' HTTP/1.1\r\n' +
    'Host: ' + new URI(BASE_URL, false).getHost() + '\r\n' +
    'Content-Type: application/json\r\n' +
    'Accept: application/json'
  );
  const payload = JSON.stringify({ email: EMAIL, password: PASSWORD });

  const msg = new HttpMessage();
  msg.setRequestHeader(reqHeader);
  msg.setRequestBody(payload);
  msg.getRequestHeader().setContentLength(msg.getRequestBody().length());
  msg.getRequestHeader().setURI(new URI(loginUrl, false));
  helper.getHttpSender().sendAndReceive(msg, true);

  const code = msg.getResponseHeader().getStatusCode();
  const body = msg.getResponseBody().toString();
  if (code === 200) {
    try {
      const json = JSON.parse(body);
      if (json && json.success === true && json.token) {
        ScriptVars.setGlobalVar(VAR_KEY, String(json.token));
        log('Login OK — token tersimpan.');
        return true;
      }
    } catch(e){ log('Gagal parse JSON: ' + e); }
  } else { log('Login gagal ' + code + ' Body: ' + body.substring(0,200)); }
  return false;
}

function sendingRequest(msg, initiator, helper){
  try{
    const uri = String(msg.getRequestHeader().getURI());
    if (!uri.startsWith(BASE_URL)) return;
    if (uri.startsWith(BASE_URL + LOGIN_PATH)) return;
    let token = ScriptVars.getGlobalVar(VAR_KEY);
    if (!token || token.length < 10){ if(!doLogin(helper)) return; token = ScriptVars.getGlobalVar(VAR_KEY); }
    msg.getRequestHeader().setHeader('Authorization', 'Bearer ' + token);
  }catch(e){ log('sendingRequest error: ' + e); }
}

function responseReceived(msg, initiator, helper){
  try{
    const uri = String(msg.getRequestHeader().getURI());
    if (!uri.startsWith(BASE_URL)) return;
    const status = msg.getResponseHeader().getStatusCode();
    if (status === 401 || status === 403){ log(status+' diterima — reset token.'); ScriptVars.setGlobalVar(VAR_KEY, null); }
  }catch(e){ log('responseReceived error: ' + e); }
}
