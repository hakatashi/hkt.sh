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
  if (localStorage.getItem('token')) {
    const token = localStorage.getItem('token');
    const authData = decodeJwt(token);
    if (authData.exp * 1000 < Date.now()) {
      // token expired
      localStorage.removeItem('token');
    }
  }

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

    const token = localStorage.getItem('token');
    const body = Object.fromEntries(new FormData(adminFormEl));
    adminFormEl.children[0].disabled = true;

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