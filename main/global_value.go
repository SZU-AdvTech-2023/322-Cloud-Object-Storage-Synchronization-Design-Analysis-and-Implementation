package main

var hashMapCloud map[string][]string

var hashMapLocal map[string][]string

const CREATE_CLOUD_FOLDER = 1
const CREATE_LOCAL_FOLDER = 2
const UPLOAD_FILE = 3
const DELETE_CLOUD_FILE = 4
const DELETE_LOCAL_FILE = 5
const DELETE_CLOUD_FOLDER = 6
const DELETE_LOCAL_FOLDER = 7
const UPDATE_CLOUD_FILE = 8
const UPDATE_LOCAL_FILE = 9
const RENAME_CLOUD_FILE = 10
const RENAME_LOCAL_FILE = 11
const DOWNLOAD_FILE = 12
const RENAME_CLOUD_FOLDER = 13
const RENAME_LOCAL_FOLDER = 14

const home = "D:/cloud_synchronization_test"
