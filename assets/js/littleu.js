function copyToClipboard(element) {
    console.log('I am fucking here ... ');
    var $temp = $("<input>");
    $("body").append($temp);
    $temp.val($(element).text()).select();
    document.execCommand("copy");
    $temp.remove();
}

$(document).ready(function () {

    function copyToClipboard() {
        console.log('I am here .... ');
        /* Get the text field */
        var copyText = document.getElementById("url_clipboard");

        /* Select the text field */
        copyText.select();
        copyText.setSelectionRange(0, 99999); /*For mobile devices*/

        /* Copy the text inside the text field */
        document.execCommand("copy");

        /* Alert the copied text */
        alert("Copied the text: " + copyText.value);
    }

    // $('#alert').hide();
    // $('#alert_error').hide();

    // if ($('#date').length) {
    //     $('#date').datepicker({
    //         format: "yyyy/mm/dd",
    //         weekStart: 1,
    //         todayBtn: "linked",
    //         todayHighlight: true
    //     });
    // }

    // $(".today").click();

    // var persontypes = $('#persontypes');
    // if (persontypes.length) {
    //     $.ajax({
    //         url: '/persontypes',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 $('#persontypes').append($('<option name="' + types[i].ID + '">').append(types[i].Type));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(' + data);
    //         }
    //     });
    // }

    // var persons = $('#persons');
    // if (persons.length) {
    //     $.ajax({
    //         url: '/persons',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 $('#persons').append($('<option name="' + types[i].ID + '">').append(types[i].Name));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(' + data);
    //         }
    //     });
    // }

    // var family = $('#family');
    // if (family.length) {
    //     $.ajax({
    //         url: '/personspertype/1',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 //$('#family').append($('<li name="1" class="list-group-item">').append(types[i]));
    //                 $('#family').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/person/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Name)
    //                 )));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }

    // var friends = $('#friends');
    // if (friends.length) {
    //     $.ajax({
    //         url: '/personspertype/2',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 //$('#friends').append($('<li name="1" class="list-group-item">').append(types[i]));
    //                 $('#friends').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/person/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Name)
    //                 )));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }

    // var coworkers = $('#coworkers');
    // if (coworkers.length) {
    //     $.ajax({
    //         url: '/personspertype/3',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 // $('#coworkers').append($('<li name="1" class="list-group-item">').append(types[i]));
    //                 $('#coworkers').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/person/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Person + ' on ' + types[i].Name)
    //                 )));
    //                 console.log("Persona: " + types[i]);
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }

    // // ~~~~~~~~~~~~~~~~~~~~ Interactions ... 
    // var familyInteractions = $('#family_interactions');
    // if (familyInteractions.length) {
    //     $.ajax({
    //         url: '/familyinteractions',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 var comment = types[i].Comment;
    //                 comment = comment.substring(0, 20) + '...';
    //                 $('#family_interactions').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/interaction/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Person + ' on ' + types[i].Date + ' ' + comment)
    //                 )));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }

    // var friendInteractions = $('#friend_interactions');
    // if (friendInteractions.length) {
    //     $.ajax({
    //         url: '/friendinteractions',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 var comment = types[i].Comment;
    //                 comment = comment.substring(0, 20) + '...';
    //                 //$('#friend_interactions').append($('<li name="1" class="list-group-item">').append(types[i].Person + ' ' + comment));
    //                 $('#friend_interactions').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/interaction/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Person + ' on ' + types[i].Date + ' ' + comment)
    //                 )));
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }

    // var coworkersInteractions = $('#coworkers_interactions');
    // if (coworkersInteractions.length) {
    //     $.ajax({
    //         url: '/coworkersinteractions',
    //         type: 'GET',
    //         data: {},
    //         success: function(data) {
    //             var types = JSON.parse(data);
    //             for (i = 0; i < types.length; i++) {
    //                 var comment = types[i].Comment;
    //                 comment = comment.substring(0, 20) + '...';

    //                 $('#coworkers_interactions').append(
    //                     $('<li name="1" class="list-group-item">').append(
    //                         $('<a>').attr('href','/interaction/' + types[i].ID).append(
    //                             $('<span>').attr('class', 'badge badge-light').append(types[i].Person + ' on ' + types[i].Date + ' ' + comment)
    //                 )));
    //                 console.log("Persona: " + types[i]);
    //             }
    //         },
    //         error: function(data) {
    //             console.log('woops! :(');
    //             console.log(data);
    //         }
    //     });
    // }
    // // ~~~~~~~~~~~~~~~~~~~~ Interactions ... 

    // $('#addperson').on('submit', function(e) {

    //     var currentForm = this;
    //     e.preventDefault();
    //     var name = $('#person_name').val();
    //     var personType = $('#persontypes').find(":selected").attr('name');
    //     var everydays = $('#interacteverydays').find(":selected").val();

    //     $.ajax({
    //         url: '/addperson',
    //         type: 'POST',
    //         data: {name: name, type: personType, everydays: everydays},
    //         success: function(data) {
    //             console.log("Good");
    //             $('#person_name').val('');
    //             $("#alert").fadeTo(2000, 500).slideUp(500, function() {
    //                 $("#alert").slideUp(500);
    //             });
    //         },
    //         error: function(data) {
    //             console.log("Error!");
    //             console.log(data);
    //             $("#alert_error").fadeTo(2000, 500).slideUp(500, function() {
    //                 $("#alert_error").slideUp(500);
    //             });
    //         }
    //     });

    // });

    // $('#addinteraction').on('submit', function(e) {

    //     var currentForm = this;
    //     e.preventDefault();
    //     var text = $('#interactiontext').val();
    //     var personId = $('#persons').find(":selected").attr('name');
    //     var date = $('#date').val();

    //     $.ajax({
    //         url: '/addinteraction',
    //         type: 'POST',
    //         data: {personId: personId, comment: text, date: date},
    //         success: function(data) {
    //             $('#interactiontext').val('');
    //             $("#alert").fadeTo(2000, 500).slideUp(500, function() {
    //                 $("#alert").slideUp(500);
    //             });
    //         },
    //         error: function(data) {
    //             $("#alert_error").fadeTo(2000, 500).slideUp(500, function() {
    //                 $("#alert_error").slideUp(500);
    //             });
    //         }
    //     });

    // });

});