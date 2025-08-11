// ===== HTTP Sender: Login + Inject JWT (improved) =====
const BASE_URL   = java.lang.System.getenv("TARGET_URL")    || "https://securesimta.my.id";
const LOGIN_PATH = "/login";
const EMAIL      = java.lang.System.getenv("LOGIN_EMAIL")    || "daniel@gmail.com";
const PASSWORD   = java.lang.System.getenv("LOGIN_PASSWORD") || "P@ssword1";
const VAR_KEY    = "jwt_taruna_token";

const URI               = Java.type('org.apache.commons.httpclient.URI');
const HttpRequestHeader = Java.type('org.parosproxy.paros.network.HttpRequestHeader');
const HttpMessage       = Java.type('org.parosproxy.paros.network.HttpMessage');
const ScriptVars        = Java.type('org.zaproxy.zap.extension.script.ScriptVars');

function log(m){ print('[JWT][taruna] ' + m); }

function doLogin(helper){
  try{
    const loginUrl = BASE_URL + LOGIN_PATH;
    const reqHeader = new HttpRequestHeader(
      'POST ' + LOGIN_PATH + ' HTTP/1.1\r\n' +
      'Host: ' + new URI(BASE_URL, false).getHost() + '\r\n' +
      'Content-Type: application/json; charset=utf-8\r\n' +
      'Accept: application/json'
    );
    const msg = new HttpMessage();
    msg.setRequestHeader(reqHeader);
    msg.setRequestBody(JSON.stringify({ email: EMAIL, password: PASSWORD }));
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
    }
    log('Login gagal ' + code + ' Body: ' + body.substring(0, 200));
  } catch(e){ log('doLogin error: ' + e); }
  return false;
}

// ---- Optional: decode exp untuk refresh sebelum kedaluwarsa
function decodeExp(jwt){
  try{
    const parts = String(jwt).split('.');
    if (parts.length !== 3) return 0;
    const Base64 = Java.type('java.util.Base64');
    const payload = JSON.parse(String(new java.lang.String(Base64.getUrlDecoder().decode(parts[1]), "UTF-8")));
    return (payload.exp || 0) * 1000;
  }catch(e){ return 0; }
}

function ensureToken(helper){
  let token = ScriptVars.getGlobalVar(VAR_KEY);
  if (!token || token.length < 10){
    if (!doLogin(helper)) return null;
    token = ScriptVars.getGlobalVar(VAR_KEY);
  }
  const exp = decodeExp(token);
  if (exp && Date.now() > exp - 30*1000) { // refresh 30s sebelum expired
    log('Token hampir expired — re-login.');
    if (!doLogin(helper)) return null;
    token = ScriptVars.getGlobalVar(VAR_KEY);
  }
  return token;
}

function inScope(uri){ return String(uri).indexOf(BASE_URL) === 0; }
function isLogin(uri){ return String(uri).indexOf(BASE_URL + LOGIN_PATH) === 0; }

function addCookieToken(hdr, token){
  let ck = hdr.getHeader('Cookie') || '';
  if (!/(^|;\s*)token=/.test(ck)) {
    ck = ck ? (ck + '; token=' + token) : ('token=' + token);
    hdr.setHeader('Cookie', ck);
  }
}

// --- sebelum tiap request ---
function sendingRequest(msg, initiator, helper){
  try{
    const uri = String(msg.getRequestHeader().getURI());
    if (!inScope(uri) || isLogin(uri)) return;

    // (Opsional) batasi hanya Spider/Ajax/Active biar nggak polusi:
//  const ALLOWED = [1,3,11]; if (ALLOWED.indexOf(initiator) === -1) return;

    const token = ensureToken(helper);
    if (!token) return;

    // Header + Cookie
    msg.getRequestHeader().setHeader('Authorization', 'Bearer ' + token);
    addCookieToken(msg.getRequestHeader(), token);

    // Referer opsional
    if (!msg.getRequestHeader().getHeader('Referer')) {
      msg.getRequestHeader().setHeader('Referer', BASE_URL + '/');
    }
  }catch(e){ log('sendingRequest error: ' + e); }
}

// --- setelah response ---
function responseReceived(msg, initiator, helper){
  try{
    const uri = String(msg.getRequestHeader().getURI());
    if (!inScope(uri)) return;

    const status = msg.getResponseHeader().getStatusCode();
    const loc = String(msg.getResponseHeader().getHeader('Location') || '');

    // Unauthorized atau diarahkan ke login → reset token supaya re-login next request
    if (status === 401 || status === 403 ||
        (status >= 300 && status < 400 && (loc.indexOf('/loginusers') !== -1 || loc.indexOf('/login') !== -1))) {
      ScriptVars.setGlobalVar(VAR_KEY, null);
      log('Unauthorized/redirect login — reset token.');
    }
  }catch(e){ log('responseReceived error: ' + e); }
}
