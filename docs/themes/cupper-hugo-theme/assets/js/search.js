(function () {
  function onEvent() {
    var filter = search.value.toUpperCase();
    var list = document.getElementById("list");
    var listItems = list.getElementsByTagName("li");
    for (i = 0; i < listItems.length; i++) {
      var item = listItems[i];
      var text = item.innerText.toUpperCase();
      if (text.indexOf(filter) > -1) {
        item.style.display = "";
      } else {
        item.style.display = "none";
      }
    }
  }

  var search = document.getElementById("search");
  if (search) {
    search.addEventListener("keyup", onEvent);
  }
})();
