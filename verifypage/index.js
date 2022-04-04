const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>人机验证</title>
    <!-- Compiled and minified CSS -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css">

    <!-- Compiled and minified JavaScript -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js"></script>
</head>
<body>
<div id="main" class="container flow-text" style="margin: auto">
    <div class="row">
        <h2 class="col s12">人机验证页面</h2>
    </div>
    <div class="row">
        <p>您的验证码是: <code>{%code}</code></p>
    </div>
</div>
</body>
</html>`
const errHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>人机验证</title>
    <!-- Compiled and minified CSS -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/css/materialize.min.css">

    <!-- Compiled and minified JavaScript -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/materialize/1.0.0/js/materialize.min.js"></script>
</head>
<body>
<div id="main" class="container flow-text" style="margin: auto">
    <div class="row">
        <h2 class="col s12">人机验证页面</h2>
    </div>
    <div class="row">
        <div class="col s12">
            <p>该验证请求已经失效，需要重新发起验证</p>
            <small>错误代码: {%code}</small>
        </div>
    </div>
</div>
</body>
</html>`
const setKey = ""


addEventListener('fetch', event => {
    event.respondWith(handleRequest(event.request))
})

/**
 * Respond with hello worker text
 * @param {Request} request
 */
async function handleRequest(request) {
    let data = request.url.split("/")
    let token = ""
    if (data.length >= 4) {
        switch (data[3]) {
            case setKey:
                await FKAD_KV.put(data[4], data[5], {expirationTtl: 60 * 60})
                token = await FKAD_KV.get(data[4])
                if (token === data[5]) {
                    return new Response("OK: " + token)
                }
                return new Response("NotOK: " + token)

            default:
                token = await FKAD_KV.get(data[3])
                let body = ""
                if (token == null) {
                    body = errHtml.replace("{%code}", data[3])
                } else {
                    body = html.replace("{%code}", token)
                }
                return new Response(body, {
                    headers: {
                        'content-type': 'text/html;charset=UTF-8',
                    }
                })
        }
    }

    return new Response(JSON.stringify(data), {
        headers: {'content-type': 'text/plain'},
    })
}