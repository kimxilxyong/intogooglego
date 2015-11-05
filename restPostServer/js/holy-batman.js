/*!
 * Holy batman functions
 *
 * Copyright 1015-x, Kim AS Yong
 * Dual licensed under the MIT and GPL Version 2 licenses.
 * http://www.opensource.org/licenses/mit-license.php
 * http://www.gnu.org/licenses/gpl-2.0.html
 *
 * @author Kim AS Yong (KimxIlxYong)
 * @version 1.0
 * @requires jQuery v2+
 * @preserve
 */
  "use strict";

	var debugLevel = 2;

  // Realm
  var realm = "holyGrailTokenGrailholy";

	// consts
	var HTTP_STATUS_REQUESTTIMEOUT = 408;
  var ERROR_NO_ERROR = 0;

	// global vars
	var newContentRequestRunning = false;
	var newContentRequestFinished	= false;
	var handShakeReadyForNextJson = true;
	var newContentRequestError = "";

  // list of loaded posts / items indexed by Post Id
  var itemHashMap = new Array();
  var errorHashMap = new Array();

  // user info
  var User = {
    Name: "",
    Level: 0,

  };
  // loading consts
  var LOADINGLIMIT = 2;

  var paramOffset = 0;  // where to start to get new rows
	var paramLimit = LOADINGLIMIT; // how many rows to get in one json request
	var LastSucessfullOffset = 0; // the offset used in the last request which returned rows
	var lastRecordCount = 0;  // is the number of comments returned, if 0 its probably the end of the thread
	var pageWasRefreshed = true;
  var runInterval = 0;
  var PeronalOPAlreadyRendered = false;
  var addContentTimeout = 5000;  // timeout for loading from json in milliseconds
  var loginTimeout = 5000;     // timeout for the login post

  var statusDogId = 0;  // Id of the status watchdog function interval
  var statusDogInterval = 1000;
  var autoLoadBackoff = 0; // Slow down autoload if at end of thread

	var jsonErrorCode = 0;
	var jsonErrorMessage = "";
	var jsonRenderedResult = "";
	var jsonActive = false;
  var lastRequestDuration = -1;

  if (localStorage.getItem("jsonSortByField") === "undefined") {
    localStorage.setItem("jsonSortByField", "commentcount"); // commentcount, score or postdate
  };

  var lastSucessfullOffsetCacheItem = "LastSucessfullOffsetCache";
  if (sessionStorage.getItem(lastSucessfullOffsetCacheItem) === "undefined") {
    sessionStorage.setItem(lastSucessfullOffsetCacheItem, LastSucessfullOffset);
  };

  if (debugLevel >= 2) {
    console.log("INIT sessionStorage.getItem(lastSucessfullOffsetCacheItem): " + sessionStorage.getItem(lastSucessfullOffsetCacheItem));
  };


  // DEBUG
  $(document).ready(function () {
    if (debugLevel > 2) {
      console.log("Bind Scroll button click event: " + $('#scrollclick').html());
    };
		$( "#scrollclick" ).on( "click", function() {
      // Scroll last post into view
      console.log("Click event target: " + $("#thread #commentlist ul li:last").html() );
      scrollElementIntoParentView("#thread #commentlist ul li:last", "#thread");
      console.log("Click event target: " + $("#errorsection").html() );
      scrollElementIntoParentView("#errorsection", "#thread");
    });

    $( "#cleartokenclick" ).on( "click", function() {
      console.log("Click event cleartokenclick" );
      //localStorage.setItem(realm, "");
      localStorage.removeItem(realm);
      logToDebugWindow("token has been cleared");
    });

    $( "#loadclick" ).on( "click", function(e) {
      if (isElementInView("#thread #commentlist ul li:last")) {
        if (debugLevel > 2) {
		      console.log("IN MANUAL Scroll:" + e);
        };
				// addContent(getContentTimeout
        addContent(5000);
				//addContent();
			} else {
				if (debugLevel > 2) {
          console.log("isElementInView false:" + e);
        };
			};
    });
  });
	// DEBUG END

	   function scrollElementIntoParentView(element, parent){
         try {
           $(parent)[0].scrollIntoView(false);
           $(parent).animate({ scrollTop: $(parent).scrollTop() + $(element).offset().top - $(parent).offset().top }, { duration: 2000, easing: 'linear'});
         }
         catch (err) {
         }
       };

	   // Remove all event handlers and all timeout/intervalls
	   function stopAndClearAll() {
       $( "#scrollclick" ).off( "click");
       $( "#loadclick" ).off( "click");
       $( "#thread" ).off("scroll");
       $(" #nav-sort-by-count").off("click");
       $(" #nav-sort-by-date").off("click");
       $(" #nav-sort-by-score").off("click");
       clearInterval(statusDogId);
     };

     function LoadingJsonStatus(on) {
       var loadingHiddenClass = "hidden"
       if (on == true) {
         $(".loading-icon").removeClass(loadingHiddenClass);
       } else {
         $(".loading-icon").addClass(loadingHiddenClass);
       };
     };

function JsonGetAndRenderPosts( url, limit, offset, filteruser, timeout, alwaysFunc ) {

	handShakeReadyForNextJson = false;
	var uri = url + "?limit=" + limit + "&offset=" + offset;
  // Test FilterBy
  if (filteruser) {
    uri = uri + "&fbp=" + filteruser;
  };
  if (debugLevel >= 2) {
	  console.log("Start RenderComments: " + uri);
  };
	jsonActive = true;
	newContentRequestFinished = false;
	newContentRequestRunning = true;
	lastRecordCount = 0;
  lastRequestDuration = -1;

  LoadingJsonStatus(true);

  // Call JSON
  //jQuery.getJSON( uri )

  var jwtToken = localStorage.getItem(realm);
  if (debugLevel > 2) {
    console.log("Bearer: " + jwtToken);
  };

  $.ajax({
  url: uri,
  dataType: 'json',
  beforeSend: function (xhr) {
    xhr.setRequestHeader('Authorization', 'Bearer ' + jwtToken);
  },
  //data: data,
  //success: callback,
  timeout: timeout // milli second timeout
  })
    .done(function( data ) {
        //console.log( "JSON Data: " + json.users[ 3 ].name );

      if (debugLevel >= 2) {
        console.log("******* AJAX DONE START *******");
      };

      $("#newjwttoken").html("New Token");

		  var commentHtml = "";
		  for (var i in data.Posts) {

        // Sanity test if the post already has been added to output
        if (itemHashMap[data.Posts[i].Id]) {
            // Post already exists, add counter
            itemHashMap[data.Posts[i].Id] = itemHashMap[data.Posts[i].Id] + 1;
            console.error("PostId " + data.Posts[i].Id + " already exists " + itemHashMap[data.Posts[i].Id] + " times, ignoring post");

        } else {
          itemHashMap[data.Posts[i].Id] = 1;

          var templateMap = {title: data.Posts[i].Title, user: data.Posts[i].User, postdate: data.Posts[i].PostDate,
                             url: data.Posts[i].Url, commentcount: data.Posts[i].CommentCount,
                             postid: data.Posts[i].Id,  score: data.Posts[i].Score};
          commentHtml += getTemplateHtml("template.singlepost", templateMap);
          PeronalOPAlreadyRendered = true;
		      lastRecordCount++;
        };



 	    };
      LastSucessfullOffset = lastRecordCount;

		  if (debugLevel >= 2) {
		    console.log("JSON OK " + uri + " finished in " + data.RequestDuration + " ms");
		    //console.log("JSON output='" + output + "'");
		  };

      lastRequestDuration = data.RequestDuration;

		  jsonRenderedResult = commentHtml;
		  jsonActive = false;

	    newContentRequestRunning = false;
	    newContentRequestFinished = true

      if (debugLevel >= 2) {
        console.log("******* AJAX DONE END *******");
      };

		  return true;
    })
  .fail(function( jqxhr, textStatus, error ) {
    jsonErrorCode = jqxhr.status;
	  jsonErrorMessage = error;

    if (debugLevel >= 2) {
	    console.log("AJAX ERROR: " + error + " textStatus: " + textStatus);
      console.log("AJAX ERROR: " + jqxhr.responseJSON.error.errors[0].message);
      console.log("AJAX ERROR: " + jqxhr);
    };
    //HTTP_STATUS_REQUESTTIMEOUT
	  // Error Template
    if (error == "timeout") {
      jsonErrorCode = HTTP_STATUS_REQUESTTIMEOUT;
    } else {
    };

    if (jqxhr.responseJSON.error.errors[0]) {
      jsonErrorMessage += "<br>Domain=" + jqxhr.responseJSON.error.errors[0].domain + ", Message=" + jqxhr.responseJSON.error.errors[0].message
    };
    if (jqxhr.responseJSON.error.errors[1]) {
      jsonErrorMessage += "<br>Domain=" + jqxhr.responseJSON.error.errors[1].domain + ", Message=" + jqxhr.responseJSON.error.errors[1].message
    };
    if (jqxhr.responseJSON.error.errors[2]) {
      jsonErrorMessage += "<br>Domain=" + jqxhr.responseJSON.error.errors[2].domain + ", Message= " + jqxhr.responseJSON.error.errors[2].message
    };
    jsonErrorMessage += "<br>" + uri;

	  var templateMap = {errorcode: jsonErrorCode, errormessage: jsonErrorMessage};
	  var errorHtml = getTemplateHtml("template.loaderror", templateMap);
	  jsonRenderedResult = errorHtml;
	  jsonActive = false;
	  newContentRequestRunning = false;
	  newContentRequestFinished = true;

    //logToDebugWindow("Fetching posts from " + uri + " failed: " + error);
    var debugMessage = "Fetching posts failed: " + jqxhr.responseJSON.Error;
    if (jqxhr.responseJSON.JwtValidationMessage != "")
    {
      debugMessage += ", " + jqxhr.responseJSON.JwtValidationMessage;
    };
    debugMessage += "(" + jqxhr.responseJSON.JwtValidationCode + ")";
    logToDebugWindow(debugMessage);

    return true;
	})
	.always(alwaysFunc);
};	// end of JsonGetAndRenderComments(


  // **** INIT Section
  $(document).ready(function () {
    // Timeago settings
    $.timeago.settings.strings.minute = "1 minute";
    $.timeago.settings.strings.hour = "a hour";
    $.timeago.settings.strings.hours = "%d hours";
    $.timeago.settings.strings.month = "a month";
    $.timeago.settings.strings.year = "a year";
    $.timeago.settings.allowFuture = true;

    // set the refresh to 30 seconds instead of 60
    // problem was that the timeago jumped from "less than a minute"
    // to "2 minutes" because of the 60 refresh
    $.timeago.settings.refreshMillis = 30000;

    // set the chronos check interval to 5 seconds instead of 50 milliseconds
    // to keep the cpu load as low as possible - it runs with 50 millis with
    // no hickups, but its just not neccessary for timeago
    chronos.minimumInterval(6000);

    var initMessage = "time: OK";

    function ResetAndRelaoad() {
      $("#thread #commentlist ul").empty();
      // global vars reset
      newContentRequestRunning = false;
      newContentRequestFinished	= false;
      handShakeReadyForNextJson = true;
      newContentRequestError = "";
      paramOffset = 0;  // where to start to get new rows
      paramLimit = LOADINGLIMIT; // how many rows to get in one json request
      LastSucessfullOffset = 0; // the offset used in the last request which returned rows
      lastRecordCount = 0;  // is the number of comments returned, if 0 its probably the end of the thread
      runInterval = 0;
      PeronalOPAlreadyRendered = false;
      addContentTimeout = 5000;  // timeout for loading from json in milliseconds
      jsonErrorCode = 0;
      jsonErrorMessage = "";
      jsonRenderedResult = "";
      jsonActive = false;
      lastRequestDuration = -1;

      // Delete list of loaded posts / items
      itemHashMap = new Array();
      errorHashMap = new Array();

      sessionStorage.setItem(lastSucessfullOffsetCacheItem, LastSucessfullOffset);

      addContent(addContentTimeout);

      $("#lasttoken").html("Undefined");
    };

			// Lazy Load
      // Function called by the scroll event of the main comment list
      var ScrollEvent = function (e) {
			  if (debugLevel >= 3) {
				  console.log("ScrollEvent fired:" + e);
				};
				//addContent;
				if (isElementInView("#thread #commentlist ul li:last")) {
          if (debugLevel >= 3) {
            console.log("IN Scroll, need addContent:" + e);
          };
          // addContent getContentTimeout
          addContent(addContentTimeout);
				};
			  return true;
			};

      if (debugLevel > 3) {
        console.log("Attach scroll");
      };
			// Add scroll event
			$("#thread").on("scroll", ScrollEvent);

      // Add NAV sorting click events
      $("#nav-sort-by-count").on("click", function() {
        if (debugLevel > 1) {
          console.log("#nav-sort-by-count.on click");
        };
        setSortOrder("commentcount"); // commentcount, score or postdate
        ResetAndRelaoad();
      });
      $(" #nav-sort-by-date").on("click", function() {
        if (debugLevel > 1) {
          console.log("#nav-sort-by-date.on click");
        };
        setSortOrder("postdate"); // commentcount, score or postdate
        ResetAndRelaoad();
      });
      $(" #nav-sort-by-score").on("click", function() {
        if (debugLevel > 1) {
          console.log("#nav-sort-by-score.on click");
        };
        setSortOrder("score"); // commentcount, score or postdate
        ResetAndRelaoad();
      });
      $(" #scroll-up").on("click", function() {
        if (debugLevel > 1) {
          console.log("#scroll-up click");
        };
        scrollList("li");
      });
      $(" #scroll-down").on("click", function() {
        if (debugLevel > 1) {
          console.log("#scroll-down click");
        };
        scrollList("li:last");
      });

      // User login/logout click events
      $( "#user-login-button" ).on( "click", function() {
        if (debugLevel > 1) {
          console.log("Click event user-login-button" );
        };
        if (DoJwtLogin($("#jwt-username").val(), $("#jwt-password").val(), loginTimeout) == ERROR_NO_ERROR) {

        };
      });
      $("#user-logout-button").on("click", function() {
        if (debugLevel > 1) {
          console.log("#user-logout-button click");
        };
        localStorage.removeItem(realm);
        logToDebugWindow("token has been cleared");
        ShowJwtLoginInfo();
      });

      $("#copyLastToken").on("click", function () {
          //<a href="jwt.io?value=" + data.token>Decode</a>
          console.log("Copy to JWT clicked");
          logToDebugWindow("Sending token to jwt.io for inspection");

      });

      initMessage = initMessage + " events: OK";

      function scrollList(elem) {
        scrollElementIntoParentView("#thread #commentlist ul " + elem, "#thread");
      };

      function setSortOrder(order) {
        localStorage.setItem("jsonSortByField", order); // commentcount, score or postdate
        // Clearall selected buttons
        $("nav>div").removeClass("holy-batman-button-selected");
        if (order == "commentcount") {
          $("#nav-sort-by-count").addClass("holy-batman-button-selected");
        } else if (order == "postdate") {
          $("#nav-sort-by-date").addClass("holy-batman-button-selected");
        } else if (order == "score") {
          $("#nav-sort-by-score").addClass("holy-batman-button-selected");
        };
        logToDebugWindow("Changing sort to: " + order);
      };

      setSortOrder(localStorage.getItem("jsonSortByField"));

      initMessage = initMessage + " sort: " + localStorage.getItem("jsonSortByField");

      // Function called every 1 second
			statusDogId = setInterval(function(){

        if ((lastRecordCount > 0) || (autoLoadBackoff <= 0)) {
          // Check if new records should be loaded
          if (isElementInView("#thread #commentlist ul li:last")) {
            if (debugLevel > 3) {
              console.log("InView autoLoad");
            };
            //addContent(contentLoadTimeout);
            addContent(addContentTimeout);
          };// End of autoloading

          if (lastRecordCount > 0) {
            autoLoadBackoff = 0;
          } else {
            autoLoadBackoff = 10;
          };
          console.log("autoLoadBackoff " + autoLoadBackoff + " lastRecordCount " + lastRecordCount);
  			  // Show number of comments already loaded
  			  $(".commentcount").html($(".usercomment").length);
          $(".requestduration").html(lastRequestDuration);
          $(".requestresultcount").html(lastRecordCount);

          $(".lasttoken").html(localStorage.getItem(realm));

        } else {
          autoLoadBackoff--;
        };
      }, statusDogInterval);

      initMessage = initMessage + " watchdog: OK";



  /*    var debugwindow = jQuery(".debug-window");
      debugwindow.slimScroll({
         color: '#fff',
         size: '10px',
         height: debugwindow.height() * 0.9,
         //start: top,
         alwaysVisible: true,
         position: 'right',
      //    height: '15px',
         railVisible: true,
          alwaysVisible: true
        });
*/

      jQuery('.debug-window').scrollbar();
      jQuery('#commentscroll').scrollbar();

      logToDebugWindow(initMessage + " INIT finished");

      logToDebugWindow("SUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXTSUPER LONG TEXT");
		});
    // **** INIT Section END

		$(document).ready(function () {
      // Show Jwt info
      ShowJwtLoginInfo ();
		  // addContent getContentTimeout
			addContent(addContentTimeout);
			return true;
		});

		function isElementInView(elem) {
      var docViewTop = $(window).scrollTop();
      var docViewBottom = docViewTop + $(window).height();
      var elemNode = $(elem);
			if (typeof elemNode === "undefined") {
			  if (debugLevel >= 2) {
			    console.error("Element " + elem + " was not found in DOM");
			  };
			  return false;
			};
      if (typeof elemNode.offset() === "undefined") {
        if (debugLevel > 2) {
          console.error("Element offset() " + elem + " was not found in DOM");
        };
        return false;
      };
      var elemTop = elemNode.offset().top;
      var elemBottom = elemTop + $(elem).height();
      return ((elemBottom <= docViewBottom) && (elemTop >= docViewTop));
    };

		function addContent(getContentTimeout) {

			// Check if a request is already running/outstanding
      if (newContentRequestRunning == true) {
        // Request alread running, exit!
        if (debugLevel > 0) {
          console.warn("Request alread running, exit! Offset " + paramOffset + " Timeout=" + getContentTimeout);
        };
        return false;
      };

			DumpStatusFlags(4, "START addContent");

      var runCount = 10;
			runInterval = getContentTimeout/runCount;
			var runIntervalLoops = 0;
			var timeoutID = 0;
			var watchDogId = 0;

			// Function to check if the JSON request has delivered, is self restarting each runInterval msecs
			var jsonWatchDog = function() {

					  DumpStatusFlags(3, "START watchdogFunc");

					  runIntervalLoops++;
					  if (newContentRequestFinished == true) {

  						DumpStatusFlags(3, "********* WATCHDOG DETECTED JSON FINISH!!!!");
  						newContentRequestFinished = false;
  						return true;
  					};
					  runCount--;
					  if (runCount <= 0) {

						clearTimeout(timeoutID);
						newContentRequestError = "Request Timeout Error: Was waiting for: " + (runInterval*runIntervalLoops) + " msecs";

						// Error Template
						var templateMap = {errorcode: HTTP_STATUS_REQUESTTIMEOUT, errormessage: newContentRequestError};
						var errorHtml = getTemplateHtml("template.loaderror", templateMap);

						$("#thread").append(errorHtml);

						scrollElementIntoParentView("#errorsection", "#thread");

					  };
					  // Watchdog if request is still running
					  if (newContentRequestRunning == true) {
			             watchDogId = setTimeout(jsonWatchDog, runInterval);
					  };
					  DumpStatusFlags(4, "END watchdogFunc");
			}; //End of watchdog

			// Start self restarting Watchdog to check if a request is still running
			//jsonWatchDog();

			// JsonAlways Callback - this event gets called if the json request finished - error or not
      var jsonAlwaysFunc = function( ) {

        DumpStatusFlags(4, "START jsonAlwaysFunc");

        var doAppendResult;
        doAppendResult = true;

        if (jsonErrorCode == 0) {
          // if no error
          if (debugLevel > 2) {
            console.log("JSON FINISHED SUCCESSFUL, next params are: active="+ jsonActive + " paramOffset=" + paramOffset + ", paramLimit=" + paramLimit);
          };
          // Check if we got back records, if no: its the end of the thread for now
          if (lastRecordCount > 0) {
            paramOffset = parseInt(paramOffset) + parseInt(paramLimit);
            sessionStorage.setItem(lastSucessfullOffsetCacheItem, paramOffset);
          };
        } else {
          console.error("JSON Error: " + jsonErrorCode + ", Msg: " + jsonErrorMessage);
          // Test if this error already has been displayed
          if (errorHashMap[jsonErrorCode]) {
            errorHashMap[jsonErrorCode] += 1;
            doAppendResult = false;
            console.error("JSON Error: " + jsonErrorCode + ", errorHashMap[jsonErrorCode]: " + errorHashMap[jsonErrorCode]);

          } else {
            errorHashMap[jsonErrorCode] = 1;
            console.error("JSON Error: " + jsonErrorCode + ", errorHashMap[] undefined");
          };
        };
        if (doAppendResult == true) {
          // ***** Append the new comments fetched from JSON server
		      $("#thread #commentlist ul").append(jsonRenderedResult);
		      // Convert date to timeago
          jQuery("abbr.timeago").timeago();
          jQuery("div.commenttimeago").timeago();
        };

        newContentRequestFinished = true;
        newContentRequestRunning = false;

		    DumpStatusFlags(4, "END jsonAlwaysFunc");
			  handShakeReadyForNextJson = true;
        LoadingJsonStatus(false);

        if (pageWasRefreshed == true) {
          // Scroll last post into view
          scrollElementIntoParentView("#thread #commentlist ul li:last", "#thread");
        };
        pageWasRefreshed = false;

			  return true;
      };
			// JsonAlways Callback END
			DumpStatusFlags(4, "START JsonGetAndRenderComments");

      if (!(sessionStorage.getItem(lastSucessfullOffsetCacheItem) === "undefined")) {
          if (pageWasRefreshed == true) {
            LastSucessfullOffset = parseInt(sessionStorage.getItem(lastSucessfullOffsetCacheItem));
            paramOffset = 0;
            if (LastSucessfullOffset > 0) {
              paramLimit = LastSucessfullOffset;
            };
            if (debugLevel >= 2) {
              console.log("INIT SETUP LOADCACHEOFFSET, next params are: active="+ jsonActive + " paramOffset=" + paramOffset + ", paramLimit=" + paramLimit);
            };
        } else {
          paramLimit = LOADINGLIMIT;
        };
      }
      // CALL AJAX
      var filterOp = $("#filter-op-username").val();
      JsonGetAndRenderPosts("/j/p/" + localStorage.getItem("jsonSortByField"), paramLimit, paramOffset, filterOp, getContentTimeout, jsonAlwaysFunc);

      DumpStatusFlags(4, "END addContent");
		};

		function getTemplateHtml(template, parameters) {
			var htmlTemplate = $(template).html();

      if (htmlTemplate == undefined) {
        console.error("getTemplateHtml: " + template + " is undefined");
        return template + " is undefined";
      }

			if (debugLevel > 3) {
			  console.log("BEFORE template " + template + ": " + htmlTemplate);
      };
			Object.keys(parameters).map(
				 function(value, index) {
					  if (debugLevel > 3) {
					    console.log( "<br>Index=" + index + ", Key=" + value + ", Data: " + parameters[value] + "<br>");
				      };
					  //errorHtml = errorHtml.replace(value, parameters[value] );
					  htmlTemplate = htmlTemplate.split("{{" + value + "}}").join( parameters[value] );

            // jQuery("abbr.timeago").timeago();
				 });
			if (debugLevel > 3) {
			  console.log("AFTER template " + template + ": " + htmlTemplate);
			};
			return htmlTemplate;
		};

    // LOGIN
    function ShowJwtLoginInfo () {
      // Get the jwt info container
      var claims =   {status: "invalid", user: -1, level: 0, exp: 134545, orig_iat: 1};
      var token = localStorage.getItem(realm);
      if (token == undefined) {
        console.error("Token empty");
      } else {
        var claim = token.split(".");
        if (claim[1] === undefined) {
          console.error("Claim Undefined");
        };
        // Decode the String
        claims = jQuery.parseJSON( atob(claim[1]) );

        logToDebugWindow( "Expires date: " + new Date(claims.exp * 1000) );
        logToDebugWindow( "Now date: " + new Date(Date.now()) );
        if ( (claims.exp * 1000) > (Date.now())) {
          logToDebugWindow("NOT EXPIRED");
        } else {
          claims.status = "expired";
        };


      };

      var jwtInfo = $("#jwtinfo");
      console.log("jWT HTML: " + jwtInfo.html());

      if (jwtInfo.html() == "") {
        var templateMap = {status: claims.status, user: claims.id, level: claims.UserLevel,
                           expires: (new Date(claims.exp * 1000)).toISOString(),
                           issuedat: (new Date(claims.orig_iat * 1000)).toISOString()};
        jwtInfo.html(getTemplateHtml("template.jwtinfo", templateMap));
      };

      var allJwtInfos = jwtInfo.find("div span");
      var attrId = "";
      allJwtInfos.each(function( index ) {
        console.log( index + ": " + $( this ).html() + ", attr: " + $(this).attr("id"));

        attrId = $(this).attr("id");
        if (attrId == "status") {
          $(this).text(claims.status);
        } else if (attrId == "user") {
          if (token == undefined) {
            $(this).text("unknown");
          } else {
            $(this).text(claims.user);
          };
        } else if (attrId == "level") {
          if (token == undefined) {
            $(this).addClass("nodisplay");
          } else {
            $(this).text(claims.level);
            $(this).removeClass("nodisplay");
          };
        } else if (attrId == "expires") {
          if (token == undefined) {
            $(this).timeago("dispose");
            $(this).addClass("nodisplay");
          } else {
            $(this).timeago("init");
            $(this).removeClass("nodisplay");
          };
        } else if (attrId == "issuedat") {
          if (token == undefined) {
            $(this).timeago("dispose");
            $(this).addClass("nodisplay");
          } else {
            $(this).timeago("init");
            $(this).removeClass("nodisplay");
          };
        };

      });
    };

    function DoJwtLogin(username, password, timeout) {
      if (debugLevel > 1) {
        console.log("Starting login " + username + ": " + password);
      };
      var postData = JSON.stringify({ "username": username, "password" : password });
      if (debugLevel > 1) {
        console.log("Starting login postdate: " + postData);
      };

      $.ajax({
        type: "POST",
        //the url where you want to sent the userName and password to
        url: '/login',
        contentType:"application/json",
        dataType: 'json',
        data: postData,
        //data: data,
        //success: callback,
      timeout: loginTimeout // milli second timeout
      })
        .done(function( data ) {

    		  if (debugLevel >= 2) {
    		    console.log("Login finished: " + data.token);
            $("#jwttoken").html(data.token);
    		    //console.log("JSON output='" + output + "'");
    		  };

          localStorage.setItem(realm, data.token);
          logToDebugWindow("Login for user " + username + " succeeded");



          var claim = data.token.split(".");
          if (claim[1] === undefined) {
            console.error("Claim Undefined");

          };
          // Decode the String
          var claims = jQuery.parseJSON( atob(claim[1]) );
          console.log(claims);
          console.log(claims.exp);
          console.log(claims.UserLevel);
          console.log(claims.id);
          console.log($.timeago(new Date(claims.orig_iat * 1000))  );
          //$("#newjwttoken").text($.timeago(234234))   expires: $.timeago(claims.exp)

          var templateMap = {status: "OK", user: claims.id, level: claims.UserLevel,
                             expires: (new Date(claims.exp * 1000)).toISOString(),
                             issuedat: (new Date(claims.orig_iat * 1000)).toISOString()};

          //var templaeMap = {status: "OK", user: claims.id, level: claims.UserLevel,
          //                    expires: (new Date(claims.exp * 1000)).toISOString(), issuedat: "2012-04-23T18:25:43.511Z"};

                             //Fri Sep 11 2015 04:46:33 GMT+0200 (Mitteleuropäische Sommerzeit)
          var UserHtml = getTemplateHtml("template.jwtinfo", templateMap);
          $("#jwtinfo").html(UserHtml)

          $("#issuedat").timeago();
          $("#expires").timeago();

          var templateMap = {user: claims.username, pass: "abc"};
          //var UserHtml = $("template.top-middle");
          //console.log( UserHtml );
          var UserHtml = getTemplateHtml("template.top-middle", templateMap);

          console.log("UserHtml: %vs", UserHtml)
          $(".top-middle").html(UserHtml);
          console.log("PeronalOPAlreadyRendered: " + PeronalOPAlreadyRendered);
    		  return true;
        })
        .fail(function( jqxhr, textStatus, error ) {

          if (debugLevel >= 2) {
    	       console.log("Login ERROR: " + error + " textStatus: " + textStatus);
    	      };
            //HTTP_STATUS_REQUESTTIMEOUT
    	       // Error Template
          if (error == "timeout") {
            jsonErrorCode = HTTP_STATUS_REQUESTTIMEOUT;
          } else {
          };
          $("#jwttoken").html("Login ERROR: " + error + " textStatus: " + textStatus);

          $("#lasttoken").html("Invalid");

          logToDebugWindow("Login for user '" + username + "' failed: " + error);
          return true;
    	 })
    	  .always(function( jqxhr, textStatus, error ) {
          if (debugLevel >= 2) {
    	       console.log("Login finished: " + error + " textStatus: " + textStatus);
    	    };

        });
    };
    // LOGIN END

    // REGISTER
    function jwtregister(username, password, timeout) {
      var postData = JSON.stringify({ "username": username, "password" : password });
      if (debugLevel > 2) {
        console.log("Starting register postData: " + postData);
      };

      $.ajax({
        type: "POST",
        //the url where you want to sent the userName and password to
        url: '/register',
        contentType:"application/json",
        dataType: 'json',
        data: postData,
        //data: data,
        //success: callback,
      timeout: loginTimeout // milli second timeout
      })
        .done(function( data ) {

          logToDebugWindow("Register for user " + username + " succeeded");
          if (debugLevel > 2) {
            console.log("Register return: " + data);
          };
          return true;
        })
        .fail(function( jqxhr, textStatus, error ) {

          if (debugLevel >= 2) {
             console.log("Register ERROR: " + error + " textStatus: " + textStatus);
            };
          //HTTP_STATUS_REQUESTTIMEOUT
          if (error == "timeout") {
            jsonErrorCode = HTTP_STATUS_REQUESTTIMEOUT;
            logToDebugWindow("Register for user '" + username + "' timed out ");
          } else {
          };

          logToDebugWindow("Register for user '" + username + "' failed: " + error);
          return true;
       })
        .always(function( jqxhr, textStatus, error ) {
          if (debugLevel >= 2) {
             console.log("Register finished: " + error + " textStatus: " + textStatus);
          };
        });
    };
    // REGISTER END

		//DEBUG
    function logToDebugWindow(text) {
      //$("pre.debug-window").append(text);
      var pre = jQuery(".debug-window");
      pre.append("<br>" + text);
      pre.scrollTop( pre.prop("scrollHeight") );
    };

		function DumpStatusFlags(level, tag) {
		  if (level <= debugLevel) {
			  console.warn("**** DUMP STATUS " + level + "/" + debugLevel + " **** " + tag);
			  console.log("newContentRequestRunning=" + newContentRequestRunning);
			  console.log("newContentRequestFinished=" + newContentRequestFinished);
			  console.log("handShakeReadyForNextJson=" + handShakeReadyForNextJson);
			  console.log("jsonActive=" + jsonActive);
			  console.log("runInterval=" + runInterval);
			  console.log("jsonErrorCode=" + jsonErrorCode + " jsonErrorMessage=" + jsonErrorMessage);
			  console.log("PARAM Offset=" + paramOffset + ", Limit=" + paramLimit + ", LastSucessfullOffset=" + LastSucessfullOffset);
			  console.log("lastRecordCount=" + lastRecordCount);
			  console.log("Start LastSucessfullOffset: " + LastSucessfullOffset);
			  console.log("PeronalOPAlreadyRendered=" + PeronalOPAlreadyRendered);
			  console.warn("** END DUMP STATUS ** " + tag);
		  };
		};
