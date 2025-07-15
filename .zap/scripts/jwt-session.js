// @name jwt-session
// @contextType session
// @description JWT session handler for ZAP
// @engine ECMAScript : Oracle Nashorn

function extractWebSession(httpMessage) {
  // Ambil token dari Authorization header
  var tokenHeader = httpMessage.getRequestHeader().getHeader("Authorization");
  return tokenHeader;
}

function replaceWebSession(httpMessage, sessionToken) {
  // Tambahkan token ke header jika tersedia
  if (sessionToken != null && sessionToken.length() > 0) {
      httpMessage.getRequestHeader().setHeader("Authorization", sessionToken);
  }
}

function clearWebSessionIdentifiers() {
  // Tidak ada data yang perlu dibersihkan
}
