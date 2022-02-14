package speed

var htmlText string = `
<!DOCTYPE html>
<html lang="en" dir="ltr">

<head>

  <meta name="keywords" content="Property management, Leasing, Property evaluation, Real Estate, Home page">
  <meta name="description"
    content="Ruktaj Ltd is a property managment company that manages assets across london. We have been around for more than 20 years within this business and has been the centre of great customer approval.">
  <meta name="author" content="Ruktaj Ltd">
  <meta charset="utf-8">
  <meta content="width=device-width, initial-scale=1" name="viewport" />

  <!-- Facebook Meta Tags -->
  <meta property="og:url" content="https://ruktaj.co.uk/" />
  <meta property="og:type" content="website" />
  <meta property="og:title" content="Ruktaj - Home Page" />
  <meta property="og:description" content="Home Page - Our values and how we will manage your property assets" />
  <!-- <meta property="og:image"              content="assets/rightmove.png" /> -->
  <meta property="og:image:secure_url" content="https://ruktaj.co.uk/assets/logo.png" />
  <meta property="fb:app_id" content="306747584729149">

  <!-- Twitter Meta Tags -->
  <meta name="twitter:card" content="summary_large_image">
  <meta property="twitter:domain" content="ruktaj.co.uk">
  <meta property="twitter:url" content="https://ruktaj.co.uk/">
  <meta name="twitter:title" content="Ruktaj - Home Page">
  <meta name="twitter:description"
    content="Ruktaj Ltd is a property managment company that manages assets across london. We have been around for more than 20 years within this business and has been the centre of great customer approval.">
  <meta name="twitter:image" content="https://ruktaj.co.uk/assets/logo.png">

  <link rel="stylesheet" href="style.css">
  <link rel="wonder" href="style.css">

  <!-- <style type="text/css">
  </style>
  <link rel="preload" href="style.css" as="style" onload="this.onload=null;this.rel='stylesheet'">
  <noscript>
    <link rel="stylesheet" href="style.css">
  </noscript> -->
  <link href="https://ruktaj.co.uk/" rel="canonical">
  <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
  <link rel="manifest" href="/site.webmanifest">
  <!-- <link rel="icon" href="favicon.ico" type="image/x-icon"> -->
  <!-- <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Raleway:wght@200;300;400;500;800&family=Roboto:wght@300;400;500;700&display=swap" rel="stylesheet"> -->
  <link rel="stylesheet" href="assets/fonts/raleway/stylesheet.css">
  <link rel="stylesheet" href="assets/fonts/roboto/stylesheet.css">
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0-beta2/css/all.min.css"
    integrity="sha512-YWzhKL2whUzgiheMoBFwW8CKV4qpHQAEuvilg9FAn5VJUDwKZZxkJNuGM4XkWuk94WCrrwslk8yWNGmY1EduTA=="
    crossorigin="anonymous" referrerpolicy="no-referrer" />


  <script src="https://ajax.googleapis.com/ajax/libs/jquery/3.6.0/jquery.min.js"></script>
  <script async src="functions.js"></script>

  <!-- Global site tag (gtag.js) - Google Analytics -->
  <!-- <script async src="https://www.googletagmanager.com/gtag/js?id=G-Z5RN6610RL"></script>
  <script>
    window.dataLayer = window.dataLayer || [];
    function gtag() { dataLayer.push(arguments); }
    gtag('js', new Date());

    gtag('config', 'G-Z5RN6610RL');
  </script> -->

  <title>Ruktaj - Home Page</title>
</head>

<body>

  <div class="container">

    <div class="background-image-intro index-bg">

      <div id="bg-wrapper" class="dark-wrapper">
        <div class="nav"></div>

        <div class="intro-container">
          <div class="intro">
            <h1><span class="intro-title">
                RUKTAJ <span>property management</span>
              </span></h1>
            <p> <span>
                Specialising in managing small to medium property owners
              </span> </p>
            <button class="button-design darker-opacity" type="button" name="button"
              onclick="location.href='evaluation.html'">Get a free valuation</button>

			  <div class="waga john"></div>
			  <span class="waga"></div>
          </div>

        </div>
      </div>

    </div>




    <div class="info-container">

      <!-- <div class="info-section info-bg1 index">

        <div class="info-lg-bx">

          sfdsf
        </div>
      </div> -->


      <div id="info-index-title" class="info-section info-bg2 index">
        <div class="info-lg-bx">
          <h1>How it works</h1>
          <hr>
        </div>
      </div>

      <div id="index-info" class="info-section info-bg2 index extension">

        <div class="info-med-bx extension">
          <h1> <b>1. </b>Advertise on the biggest websites</h1>
          <p>
            We work and create professional ad listings on your behalf
          </p>
          <p>
            All relevant information can be provided by you or we can gather it as well
          </p>
          <p>
            Here's some of sites we currently use: Rightmove, Zoopla, PrimeLocation & Gumtree.
          </p>

          <img src="assets/zoopla.webp" alt="Zoopla logo" width="150" height="100" loading="lazy">
          <img src="assets/rightmove.webp" alt="Rightmove logo" width="100" height="100" loading="lazy">
          <img src="assets/gumtree.webp" alt="Gumtree logo" width="100" height="100" loading="lazy">

        </div>

        <div class="info-med-bx">
          <h1><b>2. </b>Find the right tenant</h1>
          <p>
            Choose any of our letting plans starting from
            <b id="index-intext"> £39 per month</b>,
          </p>
          <p>
            and you'll get 2 free references including a 6 year credit check,
          </p>
          <p>
            employment and landlord reference.
          </p>

        </div>


        <div class="info-med-bx">
          <h1><b>3. </b>Stay safe, stay legal</h1>
          <p>
            Create a <b id="index-intext">watertight tenancy agreement </b>and sign it securely online.
          </p>
          <p>
            We'll register the tenant's deposit and issue all your legal certificates
          </p>
          <p>
            including a landlord privacy policy.
          </p>
        </div>

        <div class="info-med-bx">
          <h1><b>4. </b>Complete peace of mind</h1>
          <p>
            If your tenant stops paying rent, the Complete Plan ensures you will
          </p>
          <p>
            Our professional team will work with you and your tenant to resolve any disputes.
            continue to receive your full rent on time every month.
          </p>
          <p>
            In the unlikely event of eviction, you will be protected with £100,000 of legal cover and
            I kept informed at every step.
          </p>


        </div>


      </div>


      <div id="info-index-servicegrid" class="info-section info-bg1 extension">

        <h1>What we provide . . . </h1>
        <div class="info-sml-bx title-separation">
          <h2>Free inventory and check-in inspection</h2>
          <h2>Rent guarantee</h2>
          <h2>Amazing customer service</h2>
        </div>
        <div class="info-sml-bx title-separation ">
          <h2>Repair and maintenance reporting</h2>
          <h2>Check-out inspection and deposit resolution</h2>
          <h2>Guaranteed tenant</h2>
        </div>
        <div class="info-sml-bx title-separation">
          <h2>Repair and maintenance coverage</h2>
          <h2>Required documents obtained e.g. Gas certificate</h2>
          <h2>Competitive pricing</h2>
        </div>

        <div class="info-lg-bx index-button">
          <button class="button-design darker-opacity index-button" type="button" name="button"
            onclick="location.href='service.html'">See all the benefits</button>
        </div>

      </div>

      <div class="info-section info-bg2">
        <div class="info-lg-bx extension">
          <h1>Have property to manage?</h1>
          <p>
            Our ethos dfsdfsfsdfsfdsdfsdsfsdfsfsfsdfsdfs
            sdfsfd
          </p>
          <p>
            Our ethos dfsdfsfsdfsfdsdfsdsfsdfsfsfsdf
          </p>
          <p>
            Our ethos dfsdfsfs
          </p>
          <p>
            <button class="button-design darker-opacity index-button" type="button" name="button"
              onclick="location.href='evaluation.html'">Get a free valuation</button>
          </p>
        </div>
      </div>

    </div>

  </div>


  <div class="footer"></div>
  </div>













</body>

</html>

`

func GetTextbig() string {
	return htmlText
}
