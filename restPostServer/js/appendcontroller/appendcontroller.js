/*!
 * JS appendcontroller class
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

  class AppendController {
      constructor(uri, radius) {
          AppendController.acVersion = "0.1";
          this.scrollelement = null;
          this.onscroll = {};
          this._active = false;
          this.uri = uri;
          this.offset = 0;
          this.limit = 2;

          // Lazy Load
          // Function called if new data is needed
          var _loadhandler = function (elem) {
            console.log("Load event: " + elem.limit)
            return true
          };
          this._loadhandler = _loadhandler

          var self = this
          // Function called by the scroll event of the scrollable element
          this._scrollhandler = function(elem) {
            console.log("Scroll event: " + self.limit)

              self._loadhandler(self)

          };

      };



      static get acVersion() {
          return !this._version ? "0.0" : this._version;
      };
      static set acVersion(val) {
        this._version = val;
      };

      string() {
        var s = "";
        s = s + this.uri + "<br>" + this.offset + "<br>" + this.limit + "<br>";
        if (this.callback) {
          s = s + this.callback(this);
        };
        return s;
      };

      stringJson() {
        return JSON.stringify(this);
      };

      // Events getter/setter
      get onscroll() {
        return this._onscroll;
      };
      set onscroll(onscroll) {
        return this._onscroll = onscroll;
      }
      // End of Events getter/setter

      // Properties getter/setter
      get active() {
        return this._active;
      };
      set active(val) {
        if (val) {

          // Attach scroll event
          this.scrollelement = $("#thread")
          // Add scroll event
          if (this.scrollelement) {
            this.scrollelement.on("scroll", this._scrollhandler);
          }
        } else {
          $("#thread").off("scroll");

        }
        this._active = val;
      }
      get offset() {
          return this._offset;
      };
      set offset(offset) {
          if (!Number.isInteger(offset))
              throw new Error("Offset must be an integer.");

          if (offset < 0)
            throw new Error("Offset must not be a negative integer.");

          this._offset = offset;
      };
      get limit() {
          return this._limit;
      };
      set limit(limit) {
          if (!Number.isInteger(limit))
              throw new Error("Offset must be an integer.");

          if (limit < 1)
            throw new Error("Limit must be a positive integer.");

          this._limit = limit;
      };
      // End of Properties getter/setter
  }
