/*
	YYYY-MM-DD HH:mm:ss ZZ
	2019-01-12 00:15:35.150358 +0000 UTC
*/

$(document).ready(function(){
  $('datetime').each(function(){
    var $this = $(this);
    console.log($this.text());
    $this.text(moment($this.text(), "YYYY-MM-DD HH:mm:ss ZZ").calendar());
    console.log($this.text());
  });
});
