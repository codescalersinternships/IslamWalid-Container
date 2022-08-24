package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const ProcessesLimit = "10"

func main() {
    if len(os.Args) < 3 {
        fmt.Fprintf(os.Stderr, "to few arguments")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "run":
        err := Run(os.Args[2], os.Args[3:])
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
        }
        os.Exit(1)
    case "child":
        err := Child(os.Args[2], os.Args[3:])
        if err != nil {
            fmt.Fprintf(os.Stderr, err.Error())
        }
        os.Exit(1)
    }
}

// Start the parent process of the container
func Run(command string, args []string) error {
    cmd := exec.Command("/proc/self/exe", append([]string{"child", command}, args...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Credential: &syscall.Credential{
            Uid: uint32(os.Getuid()),
            Gid: uint32(os.Getgid()),
        },
        Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
        Unshareflags: syscall.CLONE_NEWNS,
        UidMappings:  []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
        GidMappings:  []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
    }

    err := cmd.Run()
    if err != nil {
        return err
    }

    return nil
}

// Start a child process to execute given command
func Child(command string, args []string) error {
    fmt.Printf("Running %v with pid %d\n", append([]string{command}, args...), os.Getpid())

    // Setup the new root
    err := syscall.Sethostname([]byte("container"))
    if err != nil {
        return err
    }

    err = CreateCgroup()
    if err != nil {
        return err
    }

    err = syscall.Chroot("./rootfs")
    if err != nil {
        return err
    }

    err = os.Chdir("/")
    if err != nil {
        return err
    }

    err = syscall.Mount("proc", "proc", "proc", 0, "")
    if err != nil {
        return err
    }

    err = syscall.Mount("dev", "dev", "tmpfs", 0, "")
    if err != nil {
        return err
    }

    cmd := exec.Command(command, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err = cmd.Run()
    if err != nil {
        return err
    }

    // Clean up
    err = syscall.Unmount("/proc", 0)
    if err != nil {
        return err
    }

    err = syscall.Unmount("/dev", 0)
    if err != nil {
        return err
    }

    return nil
}

func CreateCgroup() error {
    containerPidsDir := "/sys/fs/cgroup/pids"
    os.Mkdir(containerPidsDir, 0755)

    err := os.WriteFile(path.Join(containerPidsDir, "pids.max"), []byte(ProcessesLimit), 0700)
    if err != nil {
        return err
    }

    err = os.WriteFile(path.Join(containerPidsDir, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700)
    if err != nil {
        return err
    }

    return nil
}
