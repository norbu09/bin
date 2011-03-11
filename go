#!/usr/bin/env perl

=head1 NAME

go - a password manager ... a good one

=head1 VERSION

Version 1.0

=cut

my $VERSION = 1.0;

=head1 SYNOPSIS

go [options] [<name>]

 name           config section, see config file or Config::GitLike

 Options:
   -pass|p          password file
   -help|?          brief help message
   -verbose|v       be chatty
   -man             full documentation

 config file:
   [name]
     key = value

see Config::GitLike for details.

=cut

use common::sense;
use Config::GitLike;
use Getopt::Long;
use Pod::Usage;
use Data::Dumper;
use Mac::Pasteboard;
use Shell qw/gpg/;
use IO::Prompter;

my $help;
my $man;
my $passfile = '~/Dropbox/pass/passwords.gpg';

GetOptions(
    "pass|p=s" => \$passfile,
    'help|?'   => \$help,
    "man"      => \$man,
) or pod2usage(2);
pod2usage(1) if $help;
pod2usage( -exitstatus => 0, -verbose => 2 ) if $man;

my $config_str = gpg('-q', '-d', $passfile);
die "Could not decrypt password file: $passfile\n$!"
    unless $config_str;

my $c = Config::GitLike->new(confname => '');
$c->data({});
$c->multiple({});
$c->config_files([]);
$c->parse_content(
    content  => $config_str,
    callback => sub {
        $c->define(@_, origin => '');
    },
    error    => sub {
        $c->error_callback(@_, filename => '' );
    },
);

my $dest = $ARGV[0] if $ARGV[0];

my %tmp;
$dest = prompt(
    -prompt   => "what password are you looking for? (use <TAB> to complete)\n\n: ",
    -complete => [ grep { !$tmp{$_}++ } map { $_ =~ s/\.[^\.]+$//; $_; } keys %{ { $c->dump } } ],
) unless $dest;

if ( $c->get_regexp( key => $dest ) ) {
    my $type = $c->get( key => $dest . '.type' );

    given ($type) {
        when ('http')       { open_url(); }
        when ('https')      { open_url(); }
        when ('ssh')        { open_ssh(); }
        when ('capistrano') { open_cap(); }
        when ('app')        { open_app(); }
        when ('shell')      { open_cmmd(); }
        default             { open_info(); }
    }

}
else {
    print "Could not find config for '$dest'\n";
}

sub open_url {
    my $comm = '';

    $comm .= $c->get( key => $dest . '.type' ) . '://';
    $comm .= $c->get( key => $dest . '.host' );
    open_info();
    pbcopy( $c->get( key => $dest . '.pass' ) );
    qx/open $comm/;
}

sub open_cap {
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) . "\n" );
    qx/cap deploy/;
}

sub open_ssh {
    my $comm = '';

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
    open_info();
    my $comm = $c->get( key => $dest . '.app' );
    pbcopy( $c->get( key => $dest . '.pass' ) );
    qx(open /Applications/$comm);
}

sub open_info {
    print $/;
    print "User: " . $c->get( key => $dest . '.user' ) . "\n";
    print "Pass: " . $c->get( key => $dest . '.pass' ) . "\n";
    pbcopy( $c->get( key => $dest . '.pass' ) . "\n" );
}

sub open_cmmd {
    my $cmmd = $c->get( key => $dest . '.cmmd' );
    open_info();
    print $cmmd . $/;
    system($cmmd);
}

=head1 AUTHOR

Lenz Gschwendtner, C<< <lenz@springtimesoft.com> >>

=head1 BUGS

Please report any bugs or feature requests to C<< <lenz@springtimesoft.com> >>

=head1 SUPPORT

You can find documentation for this module with the perldoc command.

    perldoc go

=head1 ACKNOWLEDGEMENTS

Thanks to Tobias Kirschstein for adding <TAB> completion and security
improvement via in memory decryption of the passfile :)

=head1 COPYRIGHT & LICENSE

Copyright 2009 Lenz Gschwendtner, springtimesoft LTD, all rights reserved.

This program is free software; you can redistribute it and/or modify it
under the same terms as Perl itself.

=cut
