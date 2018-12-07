jQuery(document).ready(function() {
  jQuery(function() {
      jQuery('.nav-prev').click(function(){
          location.href = jQuery(this).attr('href');
      });
      jQuery('.nav-next').click(function() {
          location.href = jQuery(this).attr('href');
      });
  });

  jQuery(document).keydown(function(e) {
    // prev links - left arrow key
    if(e.which == '37') {
      jQuery('.nav.nav-prev').click();
    }

    // next links - right arrow key
    if(e.which == '39') {
      jQuery('.nav.nav-next').click();
    }
  });
});
