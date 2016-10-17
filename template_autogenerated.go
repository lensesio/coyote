package main

const (
mainTemplate = `
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>Coyote Tester | Results</title>

    <!-- Angular Material style sheet -->
    <link rel="stylesheet" href="http://ajax.googleapis.com/ajax/libs/angular_material/1.1.0/angular-material.min.css">

    <!--Fontawsome-->
    <script src="https://use.fontawesome.com/fbbd91a770.js"></script>

    <style>
        body { background-color:#f4f4f4;}
        h1, h2, h3, h4 { font-weight:200 }

        /* TABLE */
        .test-name { width:200px; }
        .test-time { width: 20px; }
        .test-icon { width: 10px; }
        .test-command { font-family: courier;}
        .test-code { width: 20px; }
        .main-row { cursor: pointer; }
        .main-row:hover { background-color: #ddd;}
        table {border-spacing: 0;}
        table tbody:nth-child(odd) tr { background-color:#eee; }
        table tbody:nth-child(even) tr { background-color:#fff; }
        table tr td, th {padding:10px;}
        table thead tr { color: rgba(0,0,0,.54); font-size: 12px; font-weight: 700;white-space: nowrap; text-align:left; }
        table tbody { color: rgba(0,0,0,.87);font-size: 13px;vertical-align: middle; }
        table tbody tr { border-top: 1px rgba(0,0,0,.12) solid; }
        .td-hidden {overflow:hidden; padding: 20px;}
        .td-hidden-std {background-color: rgba(244,244,244,0.8);padding:10px;}
        .td-hidden-error {background-color: rgba(212,72,72,0.2);padding:10px;}

        /* MISC */
        .md-avatar-small { width:10px; height:30px;}
        .icon-header-button {font-size:20px;}
        .icon-status-header-passed {width:20px; margin:10px; font-size:20px; color:green}
        .icon-status-header-failed {width:20px; margin:10px; font-size:20px; color:red}
        .icon-status-passed {width:10px; color:green}
        .icon-status-failed {width:10px; color:red}
        .summary {font-size:12px; padding-right:10px;}
        .dark-background {background-color:#2b2b2b; color: #ccc;}
        .logo-section {padding-left:30px;}
        .box {color:#ccc; background-color: #242424;text-align:center; padding: 16px;}
        .execution-details {font-family:"Lucida Console", Monaco, monospace}
        .logo {  margin-bottom: 0px;}
        .summary-section {background-color:#2b2b2b;margin:-20px;}
        .md-button-custom {width:70px; height:70px; margin:10px; background-color:#3f3f3f;}
        .md-button-custom-icon {color:#ccc; size:30px;font-size:20px;}
        .progress-bar {width:100%; -webkit-animation: fullexpand 10s ease-out;}
        .footer-github {text-align:center;margin-bottom:5px;}
        .footer-landoop-img {width:20px;float: left;padding-right:5px;}
    </style>

    <script src="https://cdnjs.cloudflare.com/ajax/libs/d3/3.4.4/d3.min.js"></script>
    <script src="https://storage.googleapis.com/artifacts-landoop/d3pie.min.js"></script>
	
     
</title>

</head>
<body ng-app="CoyoteApp" ng-cloak>

<div ng-controller="MainCtrl">

    <!--Header section-->
    <div layout="row" class="dark-background logo-section">
        <div flex="5"></div>
        <div flex layout="column" layout-align="center start">
            <h1 class="logo"><img src="coyote-logo.png" height="40" style="margin-bottom: -5px;"> Coyote</h1>
            <p class="summary execution-details">{{datalist.Title}} | Executed at {{datalist.Date}}</p>
        </div>
    </div>

    <div layout="row" class="dark-background" style="padding-bottom:30px;">
        <div flex="5"></div>
        <div layout="row" flex layout-margin >
                <div flex="35" layout="column">
                    <div class="box">
                        <h4> <b>{{ percentsucc }}%</b> Passed</h4>
                        <h6 style="margin-top:-15px;"> Total tests: {{datalist.TotalTests}} </h6>
                        <progress max="100" value="{{ percentsucc }}" class="progress-bar"></progress>
                    </div>

                    <md-content style="overflow:auto;height:265px;">
                        <md-list class="md-dense box" flex >

                            <!--Todo - go to test-->
                        <md-list-item ng-repeat="d in datalist.Results" class="md-2-line" ng-click="toggleCard($index); gotoTest('testNo'+$index)">
                            <!--TODO conditional-->
                            <i class="" aria-hidden="true" style="margin-right:20px;"
							ng-class="{ 'fa fa-times icon-status-failed': d.Errors >'0', 'fa fa-check icon-status-passed': d.Errors == '0',  }"
							></i>
                            <div class="md-list-item-text" layout="column">
                                <h3 style="color:#fff;">{{d.Name}}</h3>
                                <p style="color:#ccc; padding-top: 5px;">Passed {{ d.Passed }} out of {{ d.Total }} | {{d.TotalTime | number:2}}s  </p>
                            </div>
                        </md-list-item>
                    </md-list>
                    </md-content>

                </div>

                <div layout="column" flex>
                    <div  class="box">
                        <h3>Total duration: <b>{{datalist.TotalTime | number:0}}s</b></h3>
                        <div flex id="testTimes" style="display:inline;"></div>
                    </div>
                </div>
        </div>
        <div flex="5"></div>
    </div>

    <div layout="row">
        <div flex="5"></div>
        <div flex>
            <h3>Results</h3>
            <md-card ng-repeat="test in datalist.Results" ng-init="cardIndex = $index" id="testNo{{$index}}" >
                <md-card-header>
                    <md-card-avatar layout-align="center start ">
                        <i
						style="margin-top:10px"
						ng-class="{ 'fa fa-times icon-status-failed': test.Errors > 0, 'fa fa-check icon-status-passed': test.Errors == 0,  }" aria-hidden="true"></i>
                    </md-card-avatar>
                    <md-card-header-text layout-align="center start">
                        <span class="md-title">{{test.Name}}</span>
                    </md-card-header-text>
                    <md-button class="md-icon-button" aria-label="More"  ng-click="toggleCard(cardIndex)">
                        <i  ng-hide="showcard[cardIndex]" class="fa fa-angle-double-up icon-header-button" aria-hidden="true"></i>
                        <i  ng-show="showcard[cardIndex]" class="fa fa-angle-double-down icon-header-button" aria-hidden="true"></i>
                    </md-button>
                </md-card-header>

                <md-content ng-hide="showcard[cardIndex]" >
                    <table style="width:100%;">
                        <thead>
                        <tr>
                            <th></th>
                            <th class="test-icon"><span></span></th>
                            <th class="test-name">Action</th>
                            <th class="test-time"><i class="fa fa-clock-o" aria-hidden="true"></i>  Time (sec)</th>
                            <th hide-sm hide-xs><i class="fa fa-terminal" aria-hidden="true"></i> Command</th>
                            <th class="test-code"><span>Exit code</span></th>
                        </tr>
                        </thead>
                        <tbody  ng-repeat="dtest in test.Results"  ng-init="rowIndex=$index" ng-click="toggleRow(rowIndex,   cardIndex)" class="main-row">
                        <tr>
                            <td style="width:10px;">
                                <i class="fa fa-caret-down" aria-hidden="true" ng-hide="showRow[rowIndex+''+cardIndex]"></i>
                                <i class="fa fa-caret-up" aria-hidden="true" ng-show="showRow[rowIndex+''+cardIndex]"></i>
                            </td>
                            <td>
							<i ng-class="{ 'fa fa-times icon-status-failed': dtest.Status == 'error', 'fa fa-check icon-status-passed': dtest.Status == 'ok',  }" aria-hidden="true"></i>
                            </td>
                            <td><b>{{dtest.Name}}</b></td>
                            <td> {{dtest.Time}}</td>
                            <td hide-sm hide-xs>
                                <code>
                                {{dtest.Command}}
                                </code>
                            </td>
                            <td>{{dtest.Exit}}</td>
                        </tr>
                        <tr>
                            <td colspan="6" ng-show="showRow[rowIndex+''+cardIndex]" class="td-hidden">
			
                                <h4 hide-gt-sm>Command</h4>
                                <div hide-gt-sm>
                                    <code >
										
                                        {{dtest.Command}}
                                    </code>
                                </div>

                                <h4 ng-show="dtest.Stdout != ''" >Standard Output </h4>
                                <div style="cursor:text" ng-click="$event.stopPropagation();"  ng-show="dtest.Stdout != ''" class="td-hidden-std">
								<code >
								<span ng-repeat="stdoutline in dtest.Stdout track by $index" >
									<span ng-bind-html="consoleStdout(stdoutline)" > </span><br />
								</span>
                                </code>
                                </div>

                                <h4 ng-show="dtest.Stderr != ''" >Error Log</h4>
                                <div style="cursor:text" ng-click="$event.stopPropagation();"  ng-show="dtest.Stderr != ''" class="td-hidden-error">
                                <code>
                                    {{dtest.Stderr}}
                                </code>
                                </div>
                            </td>
                        </tr>

                        
                        </tbody>
                    </table>

                    <div layout="row" layout-align="end center">
                        <p class="summary">Passed {{test.Passed}} out of {{test.Total}} | {{test.TotalTime}} seconds </p>
                    </div>
                </md-content>
            </md-card>
        </div>
        <div flex="5"></div>
    </div>


</div>

<!--Footer-->
<h6 id="#results" class="footer-github">Report Issues & Stars!</h6>
<div flex layout="row" layout-align="center center">
    <a class="github-button" href="https://github.com/Landoop/coyote/issues" data-count-api="/repos/Landoop/coyote#open_issues_count" data-count-aria-label="# issues on GitHub" aria-label="Issue Landoop/coyote on GitHub">Issue</a>
    <a class="github-button" href="https://github.com/Landoop/coyote" data-count-href="/Landoop/coyote/stargazers" data-count-api="/repos/Landoop/coyote#stargazers_count" data-count-aria-label="# stargazers on GitHub" aria-label="Star Landoop/coyote on GitHub">Star</a>
</div>

<div flex layout="row" layout-align="center center">
    <img ng-src="https://www.landoop.com/images/landoop-blue.svg" class="footer-landoop-img">
    <p style="font-size:10px;">powered by <a href="http://www.landoop.com" style="text-decoration:none;color:blue;" target="_blank">Landoop</a></p>
</div>

<!-- Angular Material requires Angular.js Libraries -->
<script src="http://ajax.googleapis.com/ajax/libs/angularjs/1.5.5/angular.min.js"></script>
<script src="http://ajax.googleapis.com/ajax/libs/angularjs/1.5.5/angular-animate.min.js"></script>
<script src="http://ajax.googleapis.com/ajax/libs/angularjs/1.5.5/angular-aria.min.js"></script>
<script src="http://ajax.googleapis.com/ajax/libs/angularjs/1.5.5/angular-messages.min.js"></script>

<!-- Angular Material Library -->
<script src="http://ajax.googleapis.com/ajax/libs/angular_material/1.1.0/angular-material.min.js"></script>
<script src="https://storage.googleapis.com/wch/ansi_up-1.3.0.js" type="text/javascript"></script>

<!--Github Buttons-->
<script async defer src="https://buttons.github.io/buttons.js"></script>

<!-- Your application bootstrap  -->
<script type="text/javascript">

angular.module('CoyoteApp', ['ngMaterial', 'ngAnimate', 'ngAria'])
.controller('MainCtrl', function ($scope, $log, $location, $anchorScroll, $sce) {

	var data = <{=( . )=}> ;	
	
	function getRandomColor() {
		var letters = '0123456789ABCDEF';
		var color = '#';
		for (var i = 0; i < 6; i++ ) {
			color += letters[Math.floor(Math.random() * 16)];
		}
		return color;
	}

	$scope.gotoTest = function(testid) {
	  $location.hash(testid);
	  $anchorScroll();
	}
	
	$scope.consoleStdout = function(stdout) {
		var ansiStdout= ansi_up.ansi_to_html(stdout);
		var trustedAnsiStdout = $sce.trustAsHtml(ansiStdout)
		return 	trustedAnsiStdout;
	}
	
	$scope.datalist = data;

	$scope.percentsucc = data.Successful / data.TotalTests * 100;
	document.title = "Coyote Tester | " + $scope.datalist.Title + " | Results";

	$scope.showcard=[]
	$scope.toggleCard = function(id) {
		$scope.showcard[id] = !$scope.showcard[id];
	}

	$scope.showRow=[]
	$scope.toggleRow = function(  rowIndex,  cardIndex) {
		$scope.showRow[rowIndex+""+cardIndex] = !$scope.showRow[rowIndex+""+cardIndex];
	}

	$scope.toggleDetail = function($index) {
	$scope.activePosition = $scope.activePosition == $index ? -1 : $index;
	};
	
	var content = [];
	angular.forEach($scope.datalist.Results, function(results2) {
		angular.forEach(results2.Results, function(results3) {
		this.push({"label": results2.Name + results3.Name  ,"value": results3.Time,"color": getRandomColor()});
		}, content);
	});

	var pie = new d3pie("testTimes", {
		 "header": {
			 "title": {
				 "text": "",
				 "color": "#fffefe",
				 "fontSize": 34,
				 "font": "sans"
		 },
			 "subtitle": {
				 "text": "",
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
			 "canvasHeight": 300,
			 "canvasWidth": "550",
			 "pieInnerRadius": "72%",
			 "pieOuterRadius": "85%"
		 },
		 "data": {
			 "sortOrder": "label-desc",
			 "smallSegmentGrouping": {
				 "enabled": true,
				 "value": 3
			 },
		 "content": content
		 },
		 "labels": {
			 "outer": {
				 "pieDistance": 25
			 },
			 "inner": {
				"hideWhenLessThanPercentage": 3
			 },
			 "mainLabel": {
				 "color": "#ffffff",
				 "fontSize": 10,
				 "pieDistance": 15,
				 "padding": 4
			 },

			 "value": {
				 "color": "#cccc43",
				 "fontSize": 10
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
				 "background": "#242424",
				 "segmentStroke": "#f6f6f6"
			 }
		 }
	});
});
</script>



</body>
</html>
`
)
