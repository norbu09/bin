#!/usr/bin/perl

use strict;
use warnings;
use Config::GitLike;
use Getopt::Long;
use Pod::Usage;
use Data::Dumper;
use feature 'switch';
use Mac::Pasteboard;
use File::Temp qw/tempfile/;

my $help;
my $man;
my $passfile = '~/Dropbox/pass/passwords.gpg';

my $dest = pop(@ARGV);

GetOptions(
    "pass|p=s" => \$passfile,
    'help|?'   => \$help,
    "man"      => \$man,
) or pod2usage(2);
pod2usage(1) if $help;
pod2usage( -exitstatus => 0, -verbose => 2 ) if $man;

my ( $fh, $tmp ) = tempfile();
my $gpg = qx/which gpg/;
chomp($gpg);
system("$gpg --decrypt -q $passfile > $tmp");
if ( !-f $tmp ) {
    die "Could not decrypt password file: $passfile\n";
}

my $c = Config::GitLike->new( confname => $tmp );
$c->load;
unlink($tmp);

if ( $c->get_regexp( key => $dest ) ) {
    my $type = $c->get( key => $dest . '.type' );

    given ($type) {
        when ('http')       { open_url(); }
        when ('https')      { open_url(); }
        when ('ssh')        { open_ssh(); }
        when ('capistrano') { open_cap(); }
        when ('app')        { open_app(); }
        when ('info')       { open_info(); }
    }

}
else {
    print "Could not find config for '$dest'\n";
}

sub open_url {
    my $comm;

    $comm .= $c->get( key => $dest . '.type' ) . '://';
    $comm .= $c->get( key => $dest . '.host' );
    print "User: " . $c->get( key => $dest . '.user' ) . "\n";
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) );
    qx/open $comm/;
}

sub open_cap {
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) . "\n" );
    qx/cap deploy/;
}

sub open_ssh {
    my $comm;

    if ( $c->get( key => $dest . '.port' ) ) {
        $comm .= '-p ' . $c->get( key => $dest . '.port' ) . ' ';
    }
    $comm .= $c->get( key => $dest . '.user' );
    $comm .= '@';
    $comm .= $c->get( key => $dest . '.host' );
    if ( $c->get( key => $dest . '.pass' ) ) {
        print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
        pbcopy( $c->get( key => $dest . '.pass' ) . "\n" );
    }
    exec("ssh $comm");
}

sub open_app {
    my $comm;

    $comm .= $c->get( key => $dest . '.app' );
    print "User: " . $c->get( key => $dest . '.user' ) . "\n";
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) );
    qx(open /Applications/$comm);
}

sub open_info {
    print "User: " . $c->get( key => $dest . '.user' ) . "\n";
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) . "\n" );
}

=head1 NAME

go - a password manager ... a good one

=head1 VERSION

Version 0.1

=head1 SYNOPSIS

go [options]

 Options:
   -pass|p          password file
   -help|?          brief help message
   -verbose|v       be chatty
   -man             full documentation

 config file:
   [command]
     key = value


=head1 AUTHOR

Lenz Gschwendtner, C<< <lenz@springtimesoft.com> >>

=head1 BUGS

Please report any bugs or feature requests to C<< <lenz@springtimesoft.com> >>

=head1 SUPPORT

You can find documentation for this module with the perldoc command.

    perldoc go

=head1 ACKNOWLEDGEMENTS


=head1 COPYRIGHT & LICENSE

Copyright 2009 Lenz Gschwendtner, springtimesoft LTD, all rights reserved.

This program is free software; you can redistribute it and/or modify it
under the same terms as Perl itself.

=cut
