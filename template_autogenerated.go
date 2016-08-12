package main

const (
mainTemplate = `<html lang="en">
    <head>
        <meta charset="utf-8">
        <title>Coyote Tester | {{ .Title }} | Results</title>

    </head>
    <body>

        <div id="testResults" style="display:inline;width:33%"></div>
        <div id="testTimes" style="display:inline;width:66%"></div>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/d3/3.4.4/d3.min.js"></script>
        <script src="https://storage.googleapis.com/artifacts-landoop/d3pie.min.js"></script>
        <script>
         var pie = new d3pie("testResults", {
             "header": {
                 "title": {
                     "text": {{ if eq .Errors 0 }}"all passed"{{ else }}"{{ .Errors }} failed"{{ end }},
                     "color": "#fffefe",
                     "fontSize": 34,
                     "font": "sans"
                 },
                 "subtitle": {
                     "text": "{{ .Title }}",
                     "color": "#E8E6E6",
                     "fontSize": 14,
                     "font": "sans"
                 },
                 "location": "pie-center",
                 "titleSubtitlePadding": 10
             },
             "footer": {
                "text": "Coyote-tester, part of Landoop™ test-suite. {{ .Date }}",
                "color": "#E8E6E6",
                "fontSize": 14,
                "font": "open sans",
                "location": "bottom-left"
             },
             "size": {
                 "canvasHeight": 375,
                 "canvasWidth": 500,
                 "pieInnerRadius": "72%",
                 "pieOuterRadius": "92%"
             },
             "data": {
                 "sortOrder": "label-desc",
                 "content": [
                     {
                         "label": "failed",
                         "value": {{ .Errors }},
                         "color": "#e21515"
                     },
                     {
                         "label": "passed",
                         "value": {{ .Successful }},
                         "color": "#64a61f"
                     }
                 ]
             },
             "labels": {
                 "outer": {
                     "format": "label-percentage1",
                     "pieDistance": 25
                 },
                 "inner": {
                     "format": "none"
                 },
                 "mainLabel": {
                     "color": "#ffffff",
                     "fontSize": 16
                 },
                 "percentage": {
                     "color": "#919191",
                     "fontSize": 16,
                     "decimalPlaces": 1
                 },
                 "value": {
                     "color": "#cccc43",
                     "fontSize": 16
                 },
                 "lines": {
                     "enabled": true,
                     "color": "#777777"
                 },
                 "truncation": {
                     "enabled": true
                 }
             },
             "effects": {
                 "pullOutSegmentOnClick": {
                     "effect": "linear",
                     "speed": 400,
                     "size": 8
                 }
             },
             "misc": {
                 "colors": {
                     "background": "#2b2b2b",
                     "segmentStroke": "#f6f6f6"
                 }
             }
         });
        </script>
        <script>
         var pie = new d3pie("testTimes", {
             "header": {
                 "title": {
                     "text": "{{ .TotalTime | printf "%.0f" }} s",
                     "color": "#fffefe",
                     "fontSize": 34,
                     "font": "sans"
             },
                 "subtitle": {
                     "text": "total time",
                     "color": "#999999",
                     "fontSize": 14,
                     "font": "sans"
                 },
                 "location": "pie-center",
                 "titleSubtitlePadding": 10
             },
             "footer": {
                "text": "",
                "color": "#999999",
                "fontSize": 10,
                "font": "open sans",
                "location": "bottom-left"
             },
             "size": {
                 "canvasHeight": 375,
                 "canvasWidth": 700,
                 "pieInnerRadius": "72%",
                 "pieOuterRadius": "85%"
             },
             "data": {
                 "sortOrder": "label-desc",
                 "smallSegmentGrouping": {
                     "enabled": true,
                     "value": 3
                 },
                 "content": [
                     {{ range $k, $v := .Results -}}
                     {{ range $ik, $iv := .Results -}}
                     {
                         "label": "{{ $v.Name }}, {{ $iv.Name }}",
                         "value": {{ $iv.Time }},
                         "color": "{{ rotateColorCharts $k $ik }}"
                     },
                     {{- end }}
                     {{- end }}
                 ]
             },
             "labels": {
                 "outer": {
                     "format": "label-percentage1",
                     "pieDistance": 25
                 },
                 "inner": {
                     "format": "none"
                 },
                 "mainLabel": {
                     "color": "#ffffff",
                     "fontSize": 12
                 },
                 "percentage": {
                     "color": "#919191",
                     "fontSize": 12,
                     "decimalPlaces": 1
                 },
                 "value": {
                     "color": "#cccc43",
                     "fontSize": 12
                 },
                 "lines": {
                     "enabled": true,
                     "color": "#777777"
                 },
                 "truncation": {
                     "enabled": true
                 }
             },
             "effects": {
                 "pullOutSegmentOnClick": {
                     "effect": "linear",
                     "speed": 400,
                     "size": 8
                 }
             },
             "misc": {
                 "colors": {
                     "background": "#2b2b2b",
                     "segmentStroke": "#f6f6f6"
                 }
             }
         });
        </script>

        <style type="text/css">
         body {
             font-family: "Open Sans", "Helvetica Neue", Helvetica, Arial;
             font-size: 14px;
             line-height: 20px;
             font-weight: 400;
             color: #3b3b3b;
             -webkit-font-smoothing: antialiased;
             font-smoothing: antialiased;
             background: #2b2b2b;
         }

         .wrapper {
             margin: 0 auto;
             padding: 40px;
             /*max-width: 800px;*/
         }

         div.ui-tooltip {
             color: red;
             border-radius: 20px;
             /*font: bold 12px "Helvetica Neue", Sans-Serif;*/
             /*text-transform: uppercase;*/
             box-shadow: 0 0 7px black;
             /*width: 400px;*/
             word-wrap: "normal";
             max-width: 900px;
         }

         .ui-tooltip-content {
             color: black;
             font: 11px Consolas, "Liberation Mono", Menlo, Courier, monospace;
             word-wrap: "normal";
             /*max-width: 900px;*/
         }

         .table {
             margin: 0 0 40px 0;
             width: 100%;
             box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
             display: table;
         }

         @media screen and (max-width: 90%) {
             .table {
                 display: block;
             }
         }

         .row {
             display: table-row;
             background: #f6f6f6;
         }

         .row:nth-of-type(odd) {
             background: #e9e9e9;
         }

         .row.header {
             font-weight: 900;
             color: #ffffff;
             background: #ea6153;
         }
         .row.green {
             background: #27ae60;
         }
         .row.blue {
             background: #2980b9;
         }
         .row.purple {
             background: #8e44ad;
         }
         .row.gray {
             background: #2c3e50;
         }
         .row.yellow {
             background: #f1c40f;
         }
         .row.orange {
             background: #d35400;
         }
         .row.turquoise {
             background: #1abc9c;
         }

         @media screen and (max-width: 90%) {
             .row {
                 padding: 8px 0;
                 display: block;
             }
         }

         .cell {
             padding: 6px 12px;
             display: table-cell;
         }

         .cell.red {
             background: #ea6153;
         }

         .cell.green {
             background: #27ae60;
         }

         .cell.skip {
             background: #2b2b2b;
         }

         .cell.center {
             text-align: center;
         }

         .cell.width12 {
             width: 12%;
             max-width: 10px;
         }
         @media screen and (max-width: 90%) {
             .cell {
                 padding: 2px 12px;
                 display: block;
             }
         }
         /* .hideContent {overflow:hidden;line-height:1em;height:2em;}
            .showContent {line-height:1em;height:auto;}
          */
        </style>

        <div class="wrapper">

            <div class="table">
            {{ range $k, $v := .Results -}}

                <div class="{{ rotateColor $k }}">
                    <div class="cell">
                        {{ $v.Name }}
                    </div>
                    <div class="cell">
                        <!--                         Status -->
                    </div>
                    <div class="cell">
                        Time (sec)
                    </div>
                    <div class="cell">
                        Exit Code
                    </div>
                    <div class="cell">
                        Command
                    </div>
                    <div class="cell width12">
                        StdOutput
                    </div>
                    <div class="cell width12">
                        StdError
                    </div>
                </div>

                {{ range $ik, $iv := $v.Results -}}
                <div class="row">
                    <div class="cell">
                        {{ $iv.Name }}
                    </div>
                    <div class="cell {{ colorStatus $iv.Status }} center">
                         {{ if eq $iv.Status "ok" }}&#10004;{{ else }}&#10006;{{ end }}
                    </div>
                    <div class="cell">
                        {{ $iv.Time | printf "%.2f" }}
                    </div>
                    <div class="cell">
                        {{ $iv.Exit | html }}
                    </div>
                    <!-- <div class="cell" style="overflow:hidden;">
                         {{ $len := len $iv.Command }}{{ if gt $len 0 }}{{ $x := splitString $iv.Command }}
                         {{ if showmore $x }}
                         <a href="" title='{{ range $x }}{{ . | html }}</br>{{ end }}'>{{ returnFirstLine $x }}</a>
                         {{ else }}
                         {{ range $x }}{{ . | html }}</br>{{ end }}
                         {{ end }}
                         {{ end }}
                         </div> -->
                    <div class="cell">
                        {{ $iv.Command | html }}
                    </div>
                    <div class="cell" style="overflow:hidden;">
                        {{ $len := len $iv.Stdout }}{{ if gt $len 2 -}}
                        <button id="trigger_{{ $k }}_{{ $ik}}" class="trigger" data-tooltip-id="{{ $k }}_{{ $ik}}" title="{{ range $iv.Stdout }}{{ . | html }}</br>{{ end }}">view</button>
                        {{ else if eq $len 2 }}
                        {{ index $iv.Stdout 0 | html }}
                        {{- end }}
                    </div>
                    <div class="cell" style="overflow:hidden;">
                        {{ $len := len $iv.Stderr }}{{ if gt $len 2 -}}
                        <button id="trigger_{{ $k }}_{{ $ik}}" class="trigger" data-tooltip-id="{{ $k }}_{{ $ik}}" title="{{ range $iv.Stderr }}{{ . | html }}</br>{{ end }}">view</button>
                        {{ else if eq $len 2 }}
                        {{ index $iv.Stderr 0 | html }}
                        {{- end }}
                    </div>
                </div>
                {{- end }}

                <div class="row">
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell" style="font-weight: bold;">
                        Passed {{ $v.Passed }} out of {{ $v.Total }}
                    </div>
                    <div class="cell" style="font-weight: bold;">
                        {{ $v.TotalTime | printf "%.2f" }} seconds
                    </div>
                </div>

                <div class="row">
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                    <div class="cell skip"></div>
                </div>
            {{- end }}
            </div>

        </div>

        <script src="https://code.jquery.com/jquery-1.12.2.min.js"></script>
        <script src="https://code.jquery.com/ui/1.11.4/jquery-ui.min.js"></script>
        <link rel="stylesheet" href="https://code.jquery.com/ui/1.11.4/themes/smoothness/jquery-ui.css">

        <script>
         $(function () {
             //show
             $(document).on('click', '.trigger', function () {
                 $(this).addClass("on");
                 $(this).tooltip({
                     items: '.trigger.on',
                     position: {
                         my: "left+30 center",
                         at: "right center",
                         collision: "flip"
                     },
                     content: function(){
                         var element = $( this );
                         return element.attr('title')
                     }
                 });
                 $(this).trigger('mouseenter');
             });
             //hide
             $(document).on('click', '.trigger.on', function () {
                 $(this).tooltip('close');
                 $(this).removeClass("on");
             });
             //prevent mouseout and other related events from firing their handlers
             $(".trigger").on('mouseout', function (e) {
                 e.stopImmediatePropagation();
             });
         });
        </script>
    </body>
</html>
`
)