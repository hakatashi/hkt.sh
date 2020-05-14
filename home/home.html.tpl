<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>hkt.sh</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="description" content="hyper-kinematic technology. shortlinks">
    <link href="//cdn.muicss.com/mui-0.10.2/css/mui.min.css" rel="stylesheet">
    <link href="//fonts.googleapis.com/css2?family=Open+Sans:wght@800&display=swap" rel="stylesheet">
    <script src="//cdn.muicss.com/mui-0.10.2/js/mui.min.js"></script>
    <style>
      body {
        background: #eee;
      }

      fieldset {
        border: none;
      }

      .mui-panel {
        overflow-x: auto;
      }

      .mui-textfield {
        vertical-align: baseline !important;
      }

      .admin-panel {
        margin: 1em 0;
      }
      .admin-panel > * {
        display: none;
      }
      .admin-panel.is-unauthorized .admin-panel-unauthorized {
        display: block;
      }
      .admin-panel.is-forbidden .admin-panel-forbidden {
        display: block;
      }
      .admin-panel.is-success .admin-panel-success {
        display: block;
      }

      h1 {
        font-family: 'Open Sans', sans-serif;
        font-weight: 800;
        font-size: 20vmin;
        line-height: 20vmin;
        text-align: center;
        text-shadow: rgba(0, 0, 0, 0.3) 0 0.04em 0.1em;
      }

      h1 > span:nth-child(3n+2) {
        color: red;
      }

      h1 > span:nth-child(3n+3) {
        color: blue;
      }
    </style>
    <script>
      const decodeJwt = (jwt) => {
        const token = jwt.split('.')[1];
        const base64 = token.replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(atob(base64));
      };

      const setState = (state) => {
        const panelEl = document.querySelector('.admin-panel');
        for (const stateClass of ['unauthorized', 'forbidden', 'success']) {
          panelEl.classList.remove(`is-${stateClass}`);
        }
        panelEl.classList.add(`is-${state}`);
      };

      document.addEventListener('DOMContentLoaded', async () => {
        const params = new URLSearchParams(location.hash.slice(1));
        history.replaceState("", document.title, location.pathname);

        if (params.has('id_token')) {
          const token = params.get('id_token');
          const authData = decodeJwt(token);

          if (authData.email !== 'hakatasiloving@gmail.com') {
            setState('forbidden');
          } else {
            localStorage.setItem('token', token);
            setState('success');
          }
        } else if (localStorage.getItem('token')) {
          setState('success');
        }

        const adminFormEl = document.querySelector('.admin-form');
        adminFormEl.addEventListener('submit', async (event) => {
          event.preventDefault();

          if (!adminFormEl.reportValidity()) {
            return;
          }

          adminFormEl.children[0].disabled = true;

          const token = localStorage.getItem('token');
          const body = Object.fromEntries(new FormData(adminFormEl));
          const res = await fetch('/admin/entry', {
            method: 'PUT',
            headers: {
              Authorization: token,
            },
            body: JSON.stringify(body),
          });
          const data = await res.json();
          console.log(data);

          adminFormEl.reset();
          adminFormEl.children[0].disabled = false;
        });
      });
    </script>
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