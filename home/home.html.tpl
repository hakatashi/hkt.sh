<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>hkt.sh</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="hyper-kinematic technology. shortlinks">
    <link href="https://cdn.muicss.com/mui-0.10.2/css/mui.min.css" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Open+Sans:wght@800&display=swap" rel="stylesheet">
    <link href="https://{{.AssetsDomain}}/favicon.png" rel="icon" type="image/png">
    <link href="https://{{.AssetsDomain}}/home.css" rel="stylesheet">
    <script src="https://cdn.muicss.com/mui-0.10.2/js/mui.min.js"></script>
    <script src="https://{{.AssetsDomain}}/home.js" defer></script>
  </head>
  <body>
    <div class="mui-container">
      <h1>
        <span>h</span><span>k</span><span>t</span><span>.</span><span>s</span><span>h</span><span>/</span>
      </h1>
      <div class="admin-panel is-unauthorized mui--text-center">
        <div class="admin-panel-unauthorized">
          <a class="mui-btn mui-btn--primary" href="https://hkt-sh-auth.auth.ap-northeast-1.amazoncognito.com/login?response_type=token&client_id={{.UserPoolId}}&scope=openid%20email&redirect_uri=https://hkt.sh/">Login</a>
        </div>
        <div class="admin-panel-forbidden mui--text-danger mui--text-headline">
          You are not hakatashi😢
        </div>
        <div class="admin-panel-success">
          <form class="admin-form mui-form--inline">
            <fieldset>
              hkt.sh/
              <div class="mui-textfield">
                <input name="Name" type="text" required>
              </div>
              →
              <div class="mui-textfield">
                <input name="Url" type="url" required>
              </div>
              <button class="mui-btn mui-btn--small mui-btn--primary">go</button>
            </fieldset>
          </form>
        </div>
      </div>
      <div class="mui-panel mui--bg-primary mui--text-light mui--z2">
        ⚠For visitors: Consider every hkt.sh/ links as ephemeral and do not use as permalink.
      </div>
      <h2>Active Endpoints</h2>
      <div class="mui-panel">
        <table class="mui-table">
          <thead>
            <tr>
              <th>Short Names</th>
              <th>Link</th>
            </tr>
          </thead>
          <tbody>
            {{range .Entries}}
              <tr>
                <td><a href="https://hkt.sh/{{.Name}}">hkt.sh/<strong>{{.Name}}</strong></a></td>
                <td><a href="{{.Url}}">{{.Url}}</a></td>
              </tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
  </body>
</html>