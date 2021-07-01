/* Expandable sections */
(function () {
  function toggle (button, target) {
    var expanded = button.getAttribute('aria-expanded') === 'true';
    button.setAttribute('aria-expanded', !expanded);
    target.hidden = !target.hidden;
  }

  var expanders = document.querySelectorAll('[data-expands]');

  Array.prototype.forEach.call(expanders, function (expander) {
    var target = document.getElementById(expander.getAttribute('data-expands'));

    expander.addEventListener('click', function () {
      toggle(expander, target);
    })
  })
}());

/* Menu button */
(function () {
  var button = document.getElementById('menu-button');
  if (button) {
    var menu = document.getElementById('patterns-list');
    button.addEventListener('click', function() {
      var expanded = this.getAttribute('aria-expanded') === 'true';
      this.setAttribute('aria-expanded', !expanded);
    })
  }
}());

/* Persist navigation scroll point */
(function () {
  window.onbeforeunload = function () {
    var patternsNav = document.getElementById('patterns-nav');
    if (patternsNav) {
      var scrollPoint = patternsNav.scrollTop;
      localStorage.setItem('scrollPoint', scrollPoint);
    }
  }

  window.addEventListener('DOMContentLoaded', function () {
    if (document.getElementById('patterns-nav')) {
      if (window.location.href.indexOf('patterns/') !== -1) {
        document.getElementById('patterns-nav').scrollTop = parseInt(localStorage.getItem('scrollPoint'));
      } else {
        document.getElementById('patterns-nav').scrollTop = 0;
      }
    }
  })
}());


  /* Add "link here" links to <h2> headings */
  (function () {
    var headings = document.querySelectorAll('main > h2');

    Array.prototype.forEach.call(headings, function (heading) {
      var id = heading.getAttribute('id');

      if (id) {
        var newHeading = heading.cloneNode(true);
        newHeading.setAttribute('tabindex', '-1');

        var container = document.createElement('div');
        container.setAttribute('class', 'h2-container');
        container.appendChild(newHeading);

        heading.parentNode.insertBefore(container, heading);

        var link = document.createElement('a');
        link.setAttribute('href', '#' + id);
        link.innerHTML = '<svg aria-hidden="true" class="link-icon" viewBox="0 0 50 50" focusable="false"> <use xlink:href="#link"></use> </svg>';

        container.appendChild(link);

        heading.parentNode.removeChild(heading);
      }
    })
  }());


/* Enable scrolling by keyboard of code samples */
(function () {
  var codeBlocks = document.querySelectorAll('pre, .code-annotated');

  Array.prototype.forEach.call(codeBlocks, function (block) {
    if (block.querySelector('code')) {
      block.setAttribute('role', 'region');
      block.setAttribute('aria-label', 'code sample');
      if (block.scrollWidth > block.clientWidth) {
        block.setAttribute('tabindex', '0');
      }
    }
  });
}());

/* Switch and persist theme */
(function () {
  var checkbox = document.getElementById('themer');

  function persistTheme(val) {
    localStorage.setItem('darkTheme', val);
  }

  function applyDarkTheme() {
    var rules = [
      '.intro-and-nav, .main-and-footer { filter: invert(100%); }',
      '* { background-color: inherit; }',
      'img:not([src*=".svg"]), .colors, iframe, .demo-container { filter: invert(100%); }'
    ];
    rules.forEach(function(rule) {
      document.styleSheets[0].insertRule(rule);
    })
  }

  function clearDarkTheme() {
    for (let i = 0; i < document.styleSheets[0].cssRules.length; i++) {
      document.styleSheets[0].deleteRule(i);
    }
  }

  checkbox.addEventListener('change', function () {
    if (this.checked) {
      applyDarkTheme();
      persistTheme('true');
    } else {
      clearDarkTheme();
      persistTheme('false');
    }
  });

  function showTheme() {
    if (localStorage.getItem('darkTheme') === 'true') {
      applyDarkTheme();
      checkbox.checked = true;
    }
  }

  function showContent() {
    document.body.style.visibility = 'visible';
    document.body.style.opacity = 1;
  }

  window.addEventListener('DOMContentLoaded', function () {
    showTheme();
    showContent();
  });

}());
