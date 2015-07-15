  function RenderComments( url ) {
  
  console.log("Start RenderComments: " + url)
  jQuery.getJSON( url )
  .done(function( data ) {
    //console.log( "JSON Data: " + json.users[ 3 ].name );
        var output="<ul>";
        //for (var i in data.Posts) {
        //    output+='<li>' + data.Posts[i].PostDate + ' ' + ' ' + '--' + '<a href="' + data.Posts[i].Url + '">' + data.Posts[i].Title + '</a>&nbsp;<img class="comment" src="/img/comment15.png" alt="Comments"> ' + data.Posts[i].CommentCount + '</li>';
        //}
		for (var i in data.Posts) {
            output+='<li><abbr class="timeago" title="' + data.Posts[i].PostDate + '">' + data.Posts[i].PostDate + '</abbr>&nbsp;--&nbsp;' + '<a href="' + data.Posts[i].Url + '">' + data.Posts[i].Title + '</a>&nbsp;<a href="/t/' + data.Posts[i].Id + '"><img class="comment" src="/img/comment15.png" alt="Comments"> ' + data.Posts[i].CommentCount + '</a></li>';
       
        output+="</ul>";
        output+="<ul>";
        
		for (var x in data.Posts[i].Comments) {
		   output+='<li>' + data.Posts[i].Comments[x].Body + '</li>';
		}
		output+="</ul>";
        
		document.getElementById("commentlist").innerHTML=output;
 	    }
		

        
		console.log("Calling timeago")
		jQuery("abbr.timeago").timeago();
		
		console.log("JSON " + url + " finished")
  })
  .fail(function( jqxhr, textStatus, error ) {
    var err = "404 " + error;
    console.log( err );
	
	err = ' <div id="error"><a href="/"><img src="/img/404.png" alt="/"></a><p>' + err + '</div>';
	document.getElementById("commentlist").innerHTML=err;
	document.getElementsByTagName("html")[0].style.backgroundColor = "#fff";
	
});
}